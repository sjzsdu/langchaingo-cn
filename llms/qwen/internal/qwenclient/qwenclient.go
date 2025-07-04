// Package qwenclient 提供了通义千问API的客户端实现
package qwenclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// 定义API端点常量
const (
	// DashScopeBaseURL 是通义千问API的基础URL
	DashScopeBaseURL = "https://dashscope.aliyuncs.com/api/v1"

	// OpenAICompatibleBaseURL 是通义千问OpenAI兼容模式的基础URL
	OpenAICompatibleBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"

	// ChatCompletionEndpoint 是聊天补全API的端点
	ChatCompletionEndpoint = "/services/aigc/text-generation/generation"

	// OpenAIChatCompletionEndpoint 是OpenAI兼容模式的聊天补全API端点
	OpenAIChatCompletionEndpoint = "/chat/completions"
)

// APIClient 定义了API客户端的接口
type APIClient interface {
	// CreateChat 发送非流式聊天请求
	CreateChat(ctx context.Context, request ChatRequest) (*ChatResponse, error)
	
	// CreateChatStream 发送流式聊天请求
	CreateChatStream(ctx context.Context, request ChatRequest) (<-chan ChatResponseChunk, <-chan error, error)
}

// ClientConfig 包含客户端配置选项
type ClientConfig struct {
	// APIKey 是通义千问API的密钥
	APIKey string

	// BaseURL 是API的基础URL
	BaseURL string

	// HTTPClient 是用于发送HTTP请求的客户端
	HTTPClient *http.Client

	// UseOpenAICompatible 表示是否使用OpenAI兼容模式
	UseOpenAICompatible bool

	// Timeout 是HTTP请求的超时时间
	Timeout time.Duration
}

// Client 是通义千问API的客户端，实现了APIClient接口
type Client struct {
	// config 是客户端配置
	config ClientConfig
}

// ClientOption 是配置Client的函数类型
type ClientOption func(*ClientConfig) error

// 验证Client实现了APIClient接口
var _ APIClient = (*Client)(nil)

// NewClient 创建一个新的通义千问API客户端
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("API密钥不能为空")
	}

	// 默认配置
	config := ClientConfig{
		APIKey:  apiKey,
		BaseURL: DashScopeBaseURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		UseOpenAICompatible: false,
		Timeout: 120 * time.Second,
	}

	// 应用选项
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}

	// 确保HTTP客户端已设置
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	// 根据兼容模式设置基础URL
	if config.UseOpenAICompatible {
		config.BaseURL = OpenAICompatibleBaseURL
	}

	return &Client{
		config: config,
	}, nil
}

// WithBaseURL 设置API的基础URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *ClientConfig) error {
		if baseURL == "" {
			baseURL = DashScopeBaseURL
		}
		c.BaseURL = baseURL
		return nil
	}
}

// WithHTTPClient 设置HTTP客户端
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *ClientConfig) error {
		c.HTTPClient = httpClient
		return nil
	}
}

// WithOpenAICompatible 设置是否使用OpenAI兼容模式
func WithOpenAICompatible(useOpenAICompatible bool) ClientOption {
	return func(c *ClientConfig) error {
		c.UseOpenAICompatible = useOpenAICompatible
		return nil
	}
}

// WithTimeout 设置HTTP请求的超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) error {
		c.Timeout = timeout
		return nil
	}
}

// CreateChat 发送非流式聊天请求
func (c *Client) CreateChat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	// 准备请求数据
	reqData, err := c.prepareRequestData(request, false)
	if err != nil {
		return nil, fmt.Errorf("准备请求数据失败: %w", err)
	}

	// 创建HTTP请求
	req, err := c.createHTTPRequest(ctx, reqData.endpoint, reqData.body)
	if err != nil {
		return nil, err
	}

	// 发送请求并获取响应
	resp, body, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	return c.parseResponse(body)
}

// requestData 包含请求相关数据
type requestData struct {
	endpoint string
	body     []byte
}

// prepareRequestData 准备请求数据
func (c *Client) prepareRequestData(request ChatRequest, streaming bool) (requestData, error) {
	var endpoint string
	var reqBody []byte
	var err error

	if c.config.UseOpenAICompatible {
		// 使用OpenAI兼容模式
		endpoint = OpenAIChatCompletionEndpoint

		// 转换为OpenAI兼容格式
		openAIReq := convertToOpenAIRequest(request)
		openAIReq.Stream = streaming
		reqBody, err = json.Marshal(openAIReq)
	} else {
		// 使用DashScope原生API
		endpoint = ChatCompletionEndpoint
		
		// 如果是流式请求，设置增量输出标志
		if streaming {
			request.Parameters.IncrementalOutput = true
		}
		
		reqBody, err = json.Marshal(request)
	}

	if err != nil {
		return requestData{}, fmt.Errorf("序列化请求失败: %w", err)
	}

	return requestData{endpoint: endpoint, body: reqBody}, nil
}

// createHTTPRequest 创建HTTP请求
func (c *Client) createHTTPRequest(ctx context.Context, endpoint string, body []byte) (*http.Request, error) {
	url := c.config.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置通用请求头
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// sendRequest 发送HTTP请求并处理响应
func (c *Client) sendRequest(req *http.Request) (*http.Response, []byte, error) {
	// 发送请求
	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, nil, fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
		}
		return nil, nil, fmt.Errorf("API错误: %s, 代码: %s", errResp.Message, errResp.Code)
	}

	return resp, body, nil
}

// parseResponse 解析响应数据
func (c *Client) parseResponse(body []byte) (*ChatResponse, error) {
	var response ChatResponse

	if c.config.UseOpenAICompatible {
		// 解析OpenAI兼容格式响应
		var openAIResp OpenAIChatResponse
		if err := json.Unmarshal(body, &openAIResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
		// 转换为统一格式
		response = convertFromOpenAIResponse(openAIResp)
	} else {
		// 解析DashScope原生响应
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return &response, nil
}

// CreateChatStream 发送流式聊天请求
func (c *Client) CreateChatStream(ctx context.Context, request ChatRequest) (<-chan ChatResponseChunk, <-chan error, error) {
	// 准备请求数据
	reqData, err := c.prepareRequestData(request, true)
	if err != nil {
		return nil, nil, fmt.Errorf("准备请求数据失败: %w", err)
	}

	// 创建HTTP请求
	req, err := c.createHTTPRequest(ctx, reqData.endpoint, reqData.body)
	if err != nil {
		return nil, nil, err
	}

	// 设置流式请求特有的请求头
	req.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, nil, fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
		}
		return nil, nil, fmt.Errorf("API错误: %s, 代码: %s", errResp.Message, errResp.Code)
	}

	// 创建通道
	responseChan := make(chan ChatResponseChunk)
	errChan := make(chan error, 1)

	// 启动goroutine处理流式响应
	go c.processStream(ctx, resp, responseChan, errChan)

	return responseChan, errChan, nil
}

// processStream 处理流式响应
func (c *Client) processStream(ctx context.Context, resp *http.Response, responseChan chan<- ChatResponseChunk, errChan chan<- error) {
	defer resp.Body.Close()
	defer close(responseChan)
	defer close(errChan)

	reader := bufio.NewReader(resp.Body)

	for {
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
			// 继续处理
		}

		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			errChan <- fmt.Errorf("读取流数据失败: %w", err)
			return
		}

		// 处理行数据
		chunk, err := c.processStreamLine(line)
		if err != nil {
			errChan <- err
			return
		}

		// 如果有有效的块数据，发送到通道
		if chunk != nil {
			responseChan <- *chunk
		}
	}
}

// processStreamLine 处理流式响应的单行数据
func (c *Client) processStreamLine(line string) (*ChatResponseChunk, error) {
	line = strings.TrimSpace(line)
	if line == "" || line == "data: [DONE]" {
		return nil, nil
	}

	// 解析SSE数据
	if !strings.HasPrefix(line, "data: ") {
		return nil, nil
	}

	data := strings.TrimPrefix(line, "data: ")

	if c.config.UseOpenAICompatible {
		// 解析OpenAI兼容格式
		var openAIChunk OpenAIChatResponseChunk
		if err := json.Unmarshal([]byte(data), &openAIChunk); err != nil {
			return nil, fmt.Errorf("解析流数据失败: %w", err)
		}

		// 转换为统一格式
		chunk := convertFromOpenAIChunk(openAIChunk)
		return &chunk, nil
	} else {
		// 解析DashScope原生格式
		var chunk ChatResponseChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return nil, fmt.Errorf("解析流数据失败: %w", err)
		}

		return &chunk, nil
	}
}

// 转换为OpenAI兼容请求格式
func convertToOpenAIRequest(req ChatRequest) OpenAIChatRequest {
	openAIReq := OpenAIChatRequest{
		Model:       req.Model,
		Messages:    make([]OpenAIChatMessage, 0, len(req.Input.Messages)),
		Temperature: req.Parameters.Temperature,
		MaxTokens:   req.Parameters.MaxTokens,
		TopP:        req.Parameters.TopP,
		Stream:      false,
	}

	// 转换消息
	for _, msg := range req.Input.Messages {
		openAIMsg := OpenAIChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// 处理多模态内容
		if len(msg.ContentParts) > 0 {
			openAIMsg.Content = make([]OpenAIContentPart, 0, len(msg.ContentParts))
			for _, part := range msg.ContentParts {
				if part.Type == "text" {
					openAIMsg.Content = append(openAIMsg.Content.([]OpenAIContentPart), OpenAIContentPart{
						Type: "text",
						Text: part.Text,
					})
				} else if part.Type == "image" {
					openAIMsg.Content = append(openAIMsg.Content.([]OpenAIContentPart), OpenAIContentPart{
						Type: "image_url",
						ImageURL: &OpenAIImageURL{
							URL: part.ImageURL.URL,
						},
					})
				}
			}
		}

		openAIReq.Messages = append(openAIReq.Messages, openAIMsg)
	}

	// 处理工具调用
	if len(req.Parameters.Tools) > 0 {
		openAIReq.Tools = make([]OpenAITool, 0, len(req.Parameters.Tools))
		for _, tool := range req.Parameters.Tools {
			openAIReq.Tools = append(openAIReq.Tools, OpenAITool{
				Type:     "function",
				Function: tool.Function,
			})
		}
	}

	// 处理工具选择
	if req.Parameters.ToolChoice != nil {
		if req.Parameters.ToolChoice.Type == "auto" {
			openAIReq.ToolChoice = "auto"
		} else if req.Parameters.ToolChoice.Type == "function" {
			openAIReq.ToolChoice = map[string]interface{}{
				"type": "function",
				"function": map[string]string{
					"name": req.Parameters.ToolChoice.Function.Name,
				},
			}
		}
	}

	return openAIReq
}

// 从OpenAI兼容响应转换为统一格式
func convertFromOpenAIResponse(resp OpenAIChatResponse) ChatResponse {
	response := ChatResponse{
		RequestID: resp.ID,
		Output:    ChatResponseOutput{},
	}

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		response.Output.Text = choice.Message.Content

		// 处理工具调用
		if len(choice.Message.ToolCalls) > 0 {
			response.Output.ToolCalls = make([]ToolCall, 0, len(choice.Message.ToolCalls))
			for _, tc := range choice.Message.ToolCalls {
				toolCall := ToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
				response.Output.ToolCalls = append(response.Output.ToolCalls, toolCall)
			}
		}
	}

	return response
}

// 从OpenAI兼容流式响应转换为统一格式
func convertFromOpenAIChunk(chunk OpenAIChatResponseChunk) ChatResponseChunk {
	responseChunk := ChatResponseChunk{
		RequestID: chunk.ID,
		Output:    ChatResponseChunkOutput{},
	}

	if len(chunk.Choices) > 0 {
		choice := chunk.Choices[0]

		// 处理增量内容
		if choice.Delta.Content != "" {
			responseChunk.Output.Text = choice.Delta.Content
		}

		// 处理工具调用
		if len(choice.Delta.ToolCalls) > 0 {
			responseChunk.Output.ToolCalls = make([]ToolCall, 0, len(choice.Delta.ToolCalls))
			for _, tc := range choice.Delta.ToolCalls {
				toolCall := ToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
				responseChunk.Output.ToolCalls = append(responseChunk.Output.ToolCalls, toolCall)
			}
		}
	}

	return responseChunk
}

// ErrorResponse 表示API错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
