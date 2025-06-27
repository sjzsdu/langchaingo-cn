package deepseekclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	TokenEnvVarName   = "DEEPSEEK_API_KEY"  //nolint:gosec
	BaseURLEnvVarName = "DEEPSEEK_BASE_URL" //nolint:gosec
	ModelEnvVarName   = "DEEPSEEK_MODEL"    //nolint:gosec
)

// Client is a client for the DeepSeek API.
type Client struct {
	// apiKey is the API key for the DeepSeek API.
	apiKey string
	// baseURL is the base URL for the DeepSeek API.
	baseURL string
	// httpClient is the HTTP client to use for requests.
	httpClient *http.Client
}

// updateToolCalls updates the tool calls in a message with new tool calls from a delta.
func (c *Client) updateToolCalls(message *ChatResponseMessage, deltaToolCalls []ToolCall) {
	// If there are no existing tool calls, initialize the slice
	if message.ToolCalls == nil {
		message.ToolCalls = make([]ToolCall, 0)
	}

	// Process each delta tool call
	for _, deltaToolCall := range deltaToolCalls {
		// Check if this is a new tool call or an update to an existing one
		var existingToolCall *ToolCall
		for i := range message.ToolCalls {
			if message.ToolCalls[i].ID == deltaToolCall.ID {
				existingToolCall = &message.ToolCalls[i]
				break
			}
		}

		// If it's a new tool call, add it to the list
		if existingToolCall == nil {
			message.ToolCalls = append(message.ToolCalls, ToolCall{
				ID:   deltaToolCall.ID,
				Type: deltaToolCall.Type,
				Function: &FunctionCall{
					Name:      deltaToolCall.Function.Name,
					Arguments: deltaToolCall.Function.Arguments,
				},
			})
		} else {
			// Update the existing tool call
			if deltaToolCall.Type != "" {
				existingToolCall.Type = deltaToolCall.Type
			}

			// Update the function if available
			if deltaToolCall.Function != nil {
				// Initialize the function if it doesn't exist
				if existingToolCall.Function == nil {
					existingToolCall.Function = &FunctionCall{}
				}

				// Update the function name if available
				if deltaToolCall.Function.Name != "" {
					existingToolCall.Function.Name = deltaToolCall.Function.Name
				}

				// Update the function arguments if available
				if deltaToolCall.Function.Arguments != "" {
					existingToolCall.Function.Arguments += deltaToolCall.Function.Arguments
				}
			}
		}
	}
}

// New creates a new DeepSeek API client.
func New(apiKey, baseURL string, httpClient *http.Client) (*Client, error) {
	if apiKey == "" {
		// 尝试从环境变量获取 API Key
		apiKey = os.Getenv(TokenEnvVarName)
		if apiKey == "" {
			return nil, fmt.Errorf("apiKey is required")
		}
	}

	if baseURL == "" {
		// 尝试从环境变量获取 Base URL
		baseURL = os.Getenv(BaseURLEnvVarName)
		if baseURL == "" {
			baseURL = "https://api.deepseek.com"
		}
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}

// CreateChat creates a chat completion.
func (c *Client) CreateChat(
	ctx context.Context,
	request ChatRequest,
) (ChatResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)

	// 处理流式请求
	if request.Stream && request.StreamingFunc != nil {
		return c.createChatStream(ctx, url, request)
	}

	// 处理非流式请求
	reqBody, err := json.Marshal(request)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return ChatResponse{}, fmt.Errorf("API error (status code %d): %s", resp.StatusCode, string(bodyBytes))
		}
		return ChatResponse{}, fmt.Errorf("API error (status code %d): %s - %s", resp.StatusCode, errResp.Error.Type, errResp.Error.Message)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return ChatResponse{}, fmt.Errorf("decoding response: %w", err)
	}

	return chatResp, nil
}

// createChatStream handles streaming chat completions.
func (c *Client) createChatStream(
	ctx context.Context,
	url string,
	request ChatRequest,
) (ChatResponse, error) {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return ChatResponse{}, fmt.Errorf("API error (status code %d): %s", resp.StatusCode, string(bodyBytes))
		}
		return ChatResponse{}, fmt.Errorf("API error (status code %d): %s - %s", resp.StatusCode, errResp.Error.Type, errResp.Error.Message)
	}

	// 处理流式响应
	reader := bufio.NewReader(resp.Body)
	var finalResponse ChatResponse
	var reasoningContent, content string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return ChatResponse{}, fmt.Errorf("reading stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || line == "data: [DONE]" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			return ChatResponse{}, fmt.Errorf("unmarshaling stream response: %w", err)
		}

		// 更新最终响应
		finalResponse.ID = streamResp.ID
		finalResponse.Model = streamResp.Model
		finalResponse.Created = streamResp.Created
		finalResponse.SystemFingerprint = streamResp.SystemFingerprint
		finalResponse.Object = "chat.completion"

		// 处理增量内容
		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]

			// 处理推理内容（仅适用于deepseek-reasoner模型）
			if choice.Delta.ReasoningContent != "" {
				reasoningContent += choice.Delta.ReasoningContent
				if request.StreamingReasoningFunc != nil {
					if err := request.StreamingReasoningFunc(ctx, []byte(choice.Delta.ReasoningContent), []byte(choice.Delta.Content)); err != nil {
						return ChatResponse{}, fmt.Errorf("streaming reasoning function error: %w", err)
					}
				}
			}

			// 处理普通内容
			if choice.Delta.Content != "" {
				content += choice.Delta.Content
				if request.StreamingFunc != nil {
					if err := request.StreamingFunc(ctx, []byte(choice.Delta.Content)); err != nil {
						return ChatResponse{}, fmt.Errorf("streaming function error: %w", err)
					}
				}
			}

			// 处理工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				// 确保最终响应中有足够的选择
				if len(finalResponse.Choices) == 0 {
					finalResponse.Choices = []ChatResponseChoice{{
						Index: choice.Index,
						Message: ChatResponseMessage{
							Role: "assistant",
						},
					}}
				}

				// 更新工具调用
				if len(choice.Delta.ToolCalls) > 0 {
					c.updateToolCalls(&finalResponse.Choices[0].Message, choice.Delta.ToolCalls)
				}
			}

			// 更新完成原因
			if choice.FinishReason != "" {
				if len(finalResponse.Choices) == 0 {
					finalResponse.Choices = []ChatResponseChoice{{
						Index: choice.Index,
					}}
				}
				finalResponse.Choices[0].FinishReason = choice.FinishReason
			}
		}
	}

	// 构建最终响应
	if len(finalResponse.Choices) == 0 {
		finalResponse.Choices = []ChatResponseChoice{{
			Index: 0,
			Message: ChatResponseMessage{
				Role:    "assistant",
				Content: content,
			},
		}}
	} else {
		finalResponse.Choices[0].Message.Content = content
	}

	// 添加推理内容（如果有）
	if reasoningContent != "" {
		finalResponse.Choices[0].Message.ReasoningContent = reasoningContent
	}

	return finalResponse, nil
}

// 这个函数已被c.updateToolCalls方法替代，保留此注释作为提醒
