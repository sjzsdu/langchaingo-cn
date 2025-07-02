package kimiclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// CreateChat 创建一个聊天请求
func (c *Client) CreateChat(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	c.setDefaults(request)

	// 如果是流式请求，使用流式处理
	if request.Stream {
		return c.createChatStream(ctx, request)
	}

	// 否则使用普通请求
	return c.createChatNormal(ctx, request)
}

// createChatNormal 创建一个普通的聊天请求
func (c *Client) createChatNormal(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	// 序列化请求
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求
	resp, err := c.do(ctx, "/chat/completions", payloadBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	// 解析响应
	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &response, nil
}

// createChatStream 创建一个流式聊天请求
func (c *Client) createChatStream(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	// 确保请求是流式的
	request.Stream = true

	// 序列化请求
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求
	resp, err := c.do(ctx, "/chat/completions", payloadBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	// 解析流式响应
	return parseStreamingChatResponse(ctx, resp, request)
}

// ChatEvent 是流式聊天事件
type ChatEvent struct {
	Response *ChatResponseChunk
	Err      error
}

// parseStreamingChatResponse 解析流式聊天响应
func parseStreamingChatResponse(ctx context.Context, r *http.Response, request *ChatRequest) (*ChatResponse, error) {
	scanner := bufio.NewScanner(r.Body)
	responseChan := make(chan ChatEvent)

	// 在goroutine中处理流式响应
	go func() {
		defer close(responseChan)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			
			// 检查是否是[DONE]消息
			if data == "[DONE]" {
				break
			}

			// 解析流式响应块
			streamPayload := &ChatResponseChunk{}
			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
			if err != nil {
				responseChan <- ChatEvent{Response: nil, Err: fmt.Errorf("解析流事件失败: %w", err)}
				return
			}
			responseChan <- ChatEvent{Response: streamPayload, Err: nil}
		}
		if err := scanner.Err(); err != nil {
			log.Println("扫描响应时出现问题:", err)
		}
	}()

	// 构建完整响应
	response := ChatResponse{
		Choices: []struct {
			Index        int         `json:"index"`
			Message      ChatMessage `json:"message"`
			FinishReason string      `json:"finish_reason"`
		}{{Message: ChatMessage{Role: "assistant", Content: ""}}},
	}

	// 处理流式响应
	var lastResponse *ChatResponseChunk
	for streamResponse := range responseChan {
		if streamResponse.Err != nil {
			return nil, streamResponse.Err
		}

		// 更新响应
		lastResponse = streamResponse.Response

		// 处理delta
		for _, choice := range streamResponse.Response.Choices {
			if content, ok := choice.Delta["content"].(string); ok && content != "" {
				// 更新内容
				currentContent := response.Choices[0].Message.Content
				if currentContent == nil {
					currentContent = ""
				}
				response.Choices[0].Message.Content = currentContent.(string) + content

				// 调用流式函数
				if request.StreamingFunc != nil {
					err := request.StreamingFunc(ctx, []byte(content))
					if err != nil {
						return nil, fmt.Errorf("流式函数返回错误: %w", err)
					}
				}
			}

			// 处理工具调用
			if toolCalls, ok := choice.Delta["tool_calls"]; ok && toolCalls != nil {
				// 这里可以添加工具调用的处理逻辑
				// 由于工具调用的处理比较复杂，这里暂时不实现
			}

			// 更新完成原因
			if choice.FinishReason != "" {
				response.Choices[0].FinishReason = choice.FinishReason
			}
		}
	}

	// 更新响应的其他字段
	if lastResponse != nil {
		response.ID = lastResponse.ID
		response.Object = lastResponse.Object
		response.Created = lastResponse.Created
		response.Model = lastResponse.Model
	}

	return &response, nil
}

// CreateChatStream 创建一个流式聊天请求，返回流式响应通道
func (c *Client) CreateChatStream(ctx context.Context, request *ChatRequest) (<-chan ChatResponseChunk, <-chan error) {
	// 确保请求是流式的
	request.Stream = true
	c.setDefaults(request)

	// 创建通道
	chunkChan := make(chan ChatResponseChunk)
	errChan := make(chan error, 1)

	// 序列化请求
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		errChan <- fmt.Errorf("序列化请求失败: %w", err)
		close(chunkChan)
		return chunkChan, errChan
	}

	// 发送请求
	resp, err := c.do(ctx, "/chat/completions", payloadBytes)
	if err != nil {
		errChan <- err
		close(chunkChan)
		return chunkChan, errChan
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		errChan <- c.decodeError(resp)
		resp.Body.Close()
		close(chunkChan)
		return chunkChan, errChan
	}

	// 在goroutine中处理流式响应
	go func() {
		defer resp.Body.Close()
		defer close(chunkChan)
		defer close(errChan)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			
			// 检查是否是[DONE]消息
			if data == "[DONE]" {
				break
			}

			// 解析流式响应块
			streamPayload := &ChatResponseChunk{}
			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
			if err != nil {
				errChan <- fmt.Errorf("解析流事件失败: %w", err)
				return
			}

			// 发送到通道
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case chunkChan <- *streamPayload:
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("扫描响应时出现问题: %w", err)
		}
	}()

	return chunkChan, errChan
}