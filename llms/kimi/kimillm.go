// Package kimi 提供了Moonshot AI的Kimi大语言模型的Go语言客户端实现
package kimi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sjzsdu/langchaingo-cn/llms/kimi/internal/kimiclient"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

// StreamingResponse 定义了流式响应的接口
type StreamingResponse interface {
	// GetChunk 获取下一个内容块
	GetChunk() (string, error)
}

const (
	// ModelKimiV1 是Kimi V1模型
	ModelKimiV1 = "moonshot-v1-8k"

	// ModelKimiV1Pro 是Kimi V1 Pro模型
	ModelKimiV1Pro = "moonshot-v1-32k"

	// ModelKimiV1Plus 是Kimi V1 Plus模型
	ModelKimiV1Plus = "moonshot-v1-128k"

	// RoleSystem 是系统角色
	RoleSystem = "system"

	// RoleUser 是用户角色
	RoleUser = "user"

	// RoleAssistant 是助手角色
	RoleAssistant = "assistant"

	// RoleTool 是工具角色
	RoleTool = "tool"
)

var (
	// ErrEmptyResponse 表示API返回了空响应
	ErrEmptyResponse = errors.New("空响应")

	// ErrMissingAPIKey 表示缺少API密钥
	ErrMissingAPIKey = errors.New("缺少API密钥，请通过参数或环境变量 KIMI_API_KEY 设置")

	// ErrMissingModel 表示缺少模型名称
	ErrMissingModel = errors.New("缺少模型名称，请通过参数或环境变量 KIMI_MODEL 设置")

	// ErrRequestFailed 表示请求失败
	ErrRequestFailed = errors.New("请求失败")
)

// LLMConfig 包含LLM的配置选项
type LLMConfig struct {
	// CallbacksHandler 是回调处理器
	CallbacksHandler callbacks.Handler

	// Model 是要使用的模型名称
	Model string

	// Temperature 控制随机性，值越高回复越随机
	Temperature float64

	// TopP 控制词汇选择的多样性
	TopP float64

	// MaxTokens 是生成的最大令牌数
	MaxTokens int
}

// LLM 是Kimi大语言模型的客户端
type LLM struct {
	// config 是LLM的配置
	config LLMConfig

	// client 是Kimi API客户端
	client *kimiclient.Client
}

// New 创建一个新的Kimi LLM客户端
func New(opts ...Option) (*LLM, error) {
	// 获取默认选项
	options := defaultOptions()

	// 应用选项
	for _, opt := range opts {
		opt(options)
	}

	// 验证必要参数
	if options.apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	if options.model == "" {
		options.model = ModelKimiV1Pro
	}

	// 创建Kimi API客户端
	clientOpts := []kimiclient.Option{}

	if options.baseURL != "" {
		clientOpts = append(clientOpts, kimiclient.WithBaseURL(options.baseURL))
	}

	if options.httpClient != nil {
		clientOpts = append(clientOpts, kimiclient.WithHTTPClient(options.httpClient))
	}

	// 创建客户端
	baseURL := options.baseURL
	if baseURL == "" {
		baseURL = "https://api.moonshot.cn/v1"
	}
	client, err := kimiclient.New(options.apiKey, options.model, baseURL, clientOpts...)
	if err != nil {
		// 提供更详细的错误信息
		return nil, fmt.Errorf("创建Kimi客户端失败: %w (请确保提供了有效的API密钥)", err)
	}

	// 创建LLM实例
	return &LLM{
		config: LLMConfig{
			CallbacksHandler: options.callbacksHandler,
			Model:            options.model,
			Temperature:      options.temperature,
			TopP:             options.topP,
			MaxTokens:        options.maxTokens,
		},
		client: client,
	}, nil
}

// Call 调用Kimi API生成文本
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	// 处理调用选项
	llmOptions := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmOptions)
	}

	// 处理回调
	callbackHandler := o.config.CallbacksHandler

	if callbackHandler != nil {
		callbackHandler.HandleLLMStart(ctx, []string{prompt})
	}

	// 构建请求
	request := kimiclient.ChatRequest{
		Model: o.config.Model,
		Messages: []kimiclient.ChatMessage{
			{
				Role:    RoleUser,
				Content: prompt,
			},
		},
		Temperature: o.config.Temperature,
		TopP:        o.config.TopP,
		MaxTokens:   o.config.MaxTokens,
	}

	// 处理JSON模式
	if llmOptions.JSONMode {
		// Kimi可能不支持JSON模式，这里可以添加相关处理
	}

	// 发送请求
	response, err := o.client.CreateChat(ctx, &request)
	if err != nil {
		if callbackHandler != nil {
			callbackHandler.HandleLLMError(ctx, err)
		}
		return "", fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	// 处理响应
	if response == nil || len(response.Choices) == 0 {
		return "", ErrEmptyResponse
	}

	// 获取文本内容
	var content string
	if contentStr, ok := response.Choices[0].Message.Content.(string); ok {
		content = contentStr
	}

	// 处理回调 - 使用HandleText代替HandleLLMEnd
	if callbackHandler != nil {
		callbackHandler.HandleText(ctx, content)
	}

	return content, nil
}

// Generate 生成多个提示的响应
func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]string, error) {
	results := make([]string, 0, len(prompts))

	for _, prompt := range prompts {
		result, err := o.Call(ctx, prompt, options...)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GenerateContent 生成内容，支持多模态输入和工具调用
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// 解析选项
	llmOptions := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmOptions)
	}

	// 处理回调
	callbackHandler := o.config.CallbacksHandler

	if callbackHandler != nil {
		callbackHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	// 转换消息格式
	kimiMessages, err := convertToKimiMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("转换消息格式失败: %w", err)
	}

	// 构建请求参数
	request := kimiclient.ChatRequest{
		Model:         o.config.Model,
		Messages:      kimiMessages,
		Temperature:   o.config.Temperature,
		TopP:          o.config.TopP,
		MaxTokens:     o.config.MaxTokens,
		Stream:        llmOptions.StreamingFunc != nil,
		StreamingFunc: llmOptions.StreamingFunc,
	}

	// 处理工具调用
	if llmOptions.Tools != nil && len(llmOptions.Tools) > 0 {
		tools, err := convertTools(llmOptions.Tools)
		if err != nil {
			return nil, fmt.Errorf("转换工具失败: %w", err)
		}
		request.Tools = tools

		// 处理工具选择
		if llmOptions.ToolChoice != nil {
			toolChoice, err := convertToolChoice(llmOptions.ToolChoice)
			if err != nil {
				return nil, fmt.Errorf("转换工具选择失败: %w", err)
			}
			request.ToolChoice = toolChoice
		}
	}

	// 发送请求
	response, err := o.client.CreateChat(ctx, &request)
	if err != nil {
		if callbackHandler != nil {
			callbackHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	// 处理响应
	if response == nil || len(response.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	// 创建响应
	contentResponse := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				StopReason: response.Choices[0].FinishReason,
			},
		},
	}

	// 处理内容
	if contentStr, ok := response.Choices[0].Message.Content.(string); ok {
		contentResponse.Choices[0].Content = contentStr
	}

	// 处理工具调用
	if response.Choices[0].Message.ToolCalls != nil {
		var toolCalls []interface{}

		// 处理不同格式的工具调用响应
		switch tc := response.Choices[0].Message.ToolCalls.(type) {
		case []interface{}:
			// 直接使用数组
			toolCalls = tc
		case map[string]interface{}:
			// 如果是单个工具调用，包装为数组
			toolCalls = []interface{}{tc}
		default:
			// 尝试通过JSON转换
			tcBytes, err := json.Marshal(response.Choices[0].Message.ToolCalls)
			if err != nil {
				return nil, fmt.Errorf("无法序列化工具调用: %w", err)
			}

			if err := json.Unmarshal(tcBytes, &toolCalls); err != nil {
				// 尝试作为单个对象解析
				var singleToolCall map[string]interface{}
				if err := json.Unmarshal(tcBytes, &singleToolCall); err != nil {
					return nil, fmt.Errorf("无法解析工具调用: %w", err)
				}
				toolCalls = []interface{}{singleToolCall}
			}
		}

		// 处理工具调用
		if len(toolCalls) > 0 {
			contentResponse.Choices[0].ToolCalls = make([]llms.ToolCall, 0, len(toolCalls))
			for _, tc := range toolCalls {
				toolCall, ok := tc.(map[string]interface{})
				if !ok {
					continue
				}

				id, _ := toolCall["id"].(string)
				typeStr, _ := toolCall["type"].(string)
				if typeStr == "" {
					// 默认为function类型
					typeStr = "function"
				}

				// 处理函数调用
				if typeStr == "function" && toolCall["function"] != nil {
					function, ok := toolCall["function"].(map[string]interface{})
					if !ok {
						continue
					}

					name, _ := function["name"].(string)

					// 处理参数，可能是字符串或对象
					var arguments string
					switch args := function["arguments"].(type) {
					case string:
						// 直接使用字符串
						arguments = args
					default:
						// 尝试将对象转换为JSON字符串
						argsBytes, err := json.Marshal(args)
						if err == nil {
							arguments = string(argsBytes)
						}
					}

					contentResponse.Choices[0].ToolCalls = append(contentResponse.Choices[0].ToolCalls, llms.ToolCall{
						ID:   id,
						Type: typeStr,
						FunctionCall: &llms.FunctionCall{
							Name:      name,
							Arguments: arguments,
						},
					})
				}
			}
		}
	}

	// 处理回调
	if callbackHandler != nil {
		// 直接使用contentResponse作为回调参数
		callbackHandler.HandleLLMGenerateContentEnd(ctx, contentResponse)
	}

	return contentResponse, nil
}

// convertToKimiMessages 将LangChain消息转换为Kimi消息
func convertToKimiMessages(messages []llms.MessageContent) ([]kimiclient.ChatMessage, error) {
	kimiMessages := make([]kimiclient.ChatMessage, 0, len(messages))

	for _, message := range messages {
		var role string
		switch message.Role {
		case llms.ChatMessageTypeSystem:
			role = RoleSystem
		case llms.ChatMessageTypeHuman:
			role = RoleUser
		case llms.ChatMessageTypeAI:
			role = RoleAssistant
		case llms.ChatMessageTypeTool:
			role = RoleTool
		default:
			return nil, fmt.Errorf("不支持的消息类型: %s", message.Role)
		}

		// 处理不同类型的内容
		if len(message.Parts) == 0 {
			return nil, fmt.Errorf("消息没有内容部分")
		}

		// 如果只有一个文本部分，直接使用字符串内容
		if len(message.Parts) == 1 {
			if textContent, ok := message.Parts[0].(llms.TextContent); ok {
				kimiMessages = append(kimiMessages, kimiclient.ChatMessage{
					Role:    role,
					Content: textContent.Text,
				})
				continue
			}

			// 处理工具调用响应
			if toolCallResponse, ok := message.Parts[0].(llms.ToolCallResponse); ok {
				// Kimi API 期望工具响应消息的 Content 字段为字符串，tool_calls 字段为数组
				kimiMessages = append(kimiMessages, kimiclient.ChatMessage{
					Role:    role,
					Content: toolCallResponse.Content,
				})
				continue
			}
			
			// 处理工具调用
			if toolCall, ok := message.Parts[0].(llms.ToolCall); ok {
				// 构建工具调用
				toolCalls := []map[string]interface{}{
					{
						"id":   toolCall.ID,
						"type": toolCall.Type,
						"function": map[string]interface{}{
							"name":      toolCall.FunctionCall.Name,
							"arguments": toolCall.FunctionCall.Arguments,
						},
					},
				}
				
				// Kimi API 要求 assistant 消息必须设置 content 和 tool_calls 字段
				kimiMessages = append(kimiMessages, kimiclient.ChatMessage{
					Role:      role,
					Content:   " ", // 设置为空格字符串而不是空字符串，确保 content 字段不为空
					ToolCalls: toolCalls,
				})
				continue
			}
		}

		// 处理多模态内容
		contents := make([]map[string]interface{}, 0, len(message.Parts))
		for _, part := range message.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				contents = append(contents, map[string]interface{}{
					"type": "text",
					"text": p.Text,
				})
			case llms.ImageURLContent:
				contents = append(contents, map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": p.URL,
					},
				})
			case llms.BinaryContent:
				contents = append(contents, map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": fmt.Sprintf("data:%s;base64,%s", p.MIMEType, base64.StdEncoding.EncodeToString(p.Data)),
					},
				})
			default:
				return nil, fmt.Errorf("不支持的内容类型: %T", part)
			}
		}

		// 如果有多个部分，使用数组内容
		if len(contents) > 0 {
			kimiMessages = append(kimiMessages, kimiclient.ChatMessage{
				Role:    role,
				Content: contents,
			})
		}
	}

	return kimiMessages, nil
}

// convertTools 将LangChain工具转换为Kimi工具
func convertTools(tools []llms.Tool) ([]kimiclient.Tool, error) {
	// 根据Moonshot AI文档，工具数量不得超过128个
	if len(tools) > 128 {
		return nil, fmt.Errorf("工具数量不得超过128个，当前: %d", len(tools))
	}

	kimiTools := make([]kimiclient.Tool, 0, len(tools))

	for _, tool := range tools {
		// 验证工具名称是否符合规范 (^[a-zA-Z_][a-zA-Z0-9-_]63$)
		if len(tool.Function.Name) > 64 {
			return nil, fmt.Errorf("工具名称长度不得超过64个字符: %s", tool.Function.Name)
		}

		// 确保parameters字段是一个object
		params, ok := tool.Function.Parameters.(map[string]interface{})
		if !ok {
			// 尝试将其他格式转换为map
			paramsBytes, err := json.Marshal(tool.Function.Parameters)
			if err != nil {
				return nil, fmt.Errorf("无法序列化工具参数: %w", err)
			}

			var paramsMap map[string]interface{}
			if err := json.Unmarshal(paramsBytes, &paramsMap); err != nil {
				return nil, fmt.Errorf("工具参数必须是一个object: %w", err)
			}

			// 确保root是一个object
			if paramsMap["type"] != "object" {
				return nil, fmt.Errorf("工具参数的root必须是一个object")
			}

			params = paramsMap
		}

		kimiTools = append(kimiTools, kimiclient.Tool{
			Type: "function",
			Function: kimiclient.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  params,
			},
		})
	}

	return kimiTools, nil
}

// convertToolChoice 将LangChain工具选择转换为Kimi工具选择
func convertToolChoice(toolChoice interface{}) (interface{}, error) {
	switch tc := toolChoice.(type) {
	case string:
		// 处理"auto"或"none"等字符串选项
		return tc, nil
	case llms.ToolChoice:
		// 处理特定工具选择
		return map[string]interface{}{
			"type": tc.Type,
			"function": map[string]interface{}{
				"name": tc.Function.Name,
			},
		}, nil
	default:
		return nil, fmt.Errorf("不支持的工具选择类型: %T", toolChoice)
	}
}

// streamingResponse 实现了StreamingResponse接口
type streamingResponse struct {
	ctx              context.Context
	chunkChan        <-chan kimiclient.ChatResponseChunk
	errChan          <-chan error
	callbacksHandler callbacks.Handler
	text             strings.Builder
}

// GetChunk 获取下一个内容块
func (s *streamingResponse) GetChunk() (string, error) {
	select {
	case <-s.ctx.Done():
		return "", s.ctx.Err()
	case err := <-s.errChan:
		if s.callbacksHandler != nil {
			s.callbacksHandler.HandleLLMError(s.ctx, err)
		}
		return "", err
	case chunk, ok := <-s.chunkChan:
		if !ok {
			// 流结束
			if s.callbacksHandler != nil {
				s.callbacksHandler.HandleText(s.ctx, s.text.String())
			}
			return "", io.EOF
		}

		// 提取内容
		var content string
		if len(chunk.Choices) > 0 {
			if contentStr, ok := chunk.Choices[0].Delta["content"].(string); ok {
				content = contentStr
			}
		}

		// 更新累积的文本
		s.text.WriteString(content)

		return content, nil
	}
}

// streamingContentResponse 实现了StreamingResponse接口，用于处理内容生成
type streamingContentResponse struct {
	ctx              context.Context
	chunkChan        <-chan kimiclient.ChatResponseChunk
	errChan          <-chan error
	callbacksHandler callbacks.Handler
	text             strings.Builder
	toolCalls        []llms.ToolCall
	currentChoice    *llms.ContentChoice
}

// GetChunk 获取下一个内容块
func (s *streamingContentResponse) GetChunk() (string, error) {
	select {
	case <-s.ctx.Done():
		return "", s.ctx.Err()
	case err := <-s.errChan:
		if s.callbacksHandler != nil {
			s.callbacksHandler.HandleLLMError(s.ctx, err)
		}
		return "", err
	case chunk, ok := <-s.chunkChan:
		if !ok {
			// 流结束
			if s.callbacksHandler != nil {
				// 创建内容响应
				contentResponse := &llms.ContentResponse{
					Choices: []*llms.ContentChoice{
						{
							Content:    s.text.String(),
							ToolCalls:  s.toolCalls,
							StopReason: s.currentChoice.StopReason,
						},
					},
				}
				s.callbacksHandler.HandleLLMGenerateContentEnd(s.ctx, contentResponse)
			}
			return "", io.EOF
		}

		// 提取内容
		var content string
		if len(chunk.Choices) > 0 {
			// 更新当前选择的停止原因
			if chunk.Choices[0].FinishReason != "" {
				s.currentChoice.StopReason = chunk.Choices[0].FinishReason
			}

			// 处理内容
			if contentStr, ok := chunk.Choices[0].Delta["content"].(string); ok {
				content = contentStr
				s.text.WriteString(content)
			}

			// 处理工具调用
			var toolCallsData []interface{}
			switch tc := chunk.Choices[0].Delta["tool_calls"].(type) {
			case []interface{}:
				// 直接使用数组
				toolCallsData = tc
			case map[string]interface{}:
				// 如果是单个工具调用，包装为数组
				toolCallsData = []interface{}{tc}
			default:
				// 尝试通过JSON转换
				if chunk.Choices[0].Delta["tool_calls"] != nil {
					tcBytes, err := json.Marshal(chunk.Choices[0].Delta["tool_calls"])
					if err == nil {
						if err := json.Unmarshal(tcBytes, &toolCallsData); err != nil {
							// 尝试作为单个对象解析
							var singleToolCall map[string]interface{}
							if err := json.Unmarshal(tcBytes, &singleToolCall); err == nil {
								toolCallsData = []interface{}{singleToolCall}
							}
						}
					}
				}
			}

			if len(toolCallsData) > 0 {
				// 初始化工具调用列表（如果需要）
				if s.toolCalls == nil {
					s.toolCalls = make([]llms.ToolCall, 0)
				}

				// 处理每个工具调用
				for _, tc := range toolCallsData {
					toolCall, ok := tc.(map[string]interface{})
					if !ok {
						continue
					}

					// 获取工具调用的索引
					var idx int
					if index, ok := toolCall["index"].(float64); ok {
						idx = int(index)
					} else {
						// 如果没有索引，使用当前长度
						idx = len(s.toolCalls)
					}

					// 确保工具调用列表有足够的空间
					for len(s.toolCalls) <= idx {
						s.toolCalls = append(s.toolCalls, llms.ToolCall{
							Type:         "function",
							FunctionCall: &llms.FunctionCall{},
						})
					}

					// 更新ID
					if id, ok := toolCall["id"].(string); ok {
						s.toolCalls[idx].ID = id
					}

					// 更新类型
					if typeStr, ok := toolCall["type"].(string); ok {
						s.toolCalls[idx].Type = typeStr
					}

					// 处理函数调用
					if function, ok := toolCall["function"].(map[string]interface{}); ok {
						// 确保 FunctionCall 已初始化
						if s.toolCalls[idx].FunctionCall == nil {
							s.toolCalls[idx].FunctionCall = &llms.FunctionCall{}
						}

						// 更新函数名称
						if name, ok := function["name"].(string); ok {
							s.toolCalls[idx].FunctionCall.Name = name
						}

						// 更新函数参数
						switch args := function["arguments"].(type) {
						case string:
							// 对于流式响应，可能需要累积参数
							s.toolCalls[idx].FunctionCall.Arguments += args
						default:
							// 尝试将对象转换为JSON字符串
							if function["arguments"] != nil {
								argsBytes, err := json.Marshal(function["arguments"])
								if err == nil {
									s.toolCalls[idx].FunctionCall.Arguments += string(argsBytes)
								}
							}
						}
					}
				}

				// 更新当前选择的工具调用
				s.currentChoice.ToolCalls = s.toolCalls
			}
		}

		return content, nil
	}
}

// StreamingCall 执行流式调用，返回流式响应接口
func (o *LLM) StreamingCall(ctx context.Context, prompt string, options ...llms.CallOption) (StreamingResponse, error) {
	// 处理调用选项
	llmOptions := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmOptions)
	}

	// 处理回调
	callbackHandler := o.config.CallbacksHandler

	if callbackHandler != nil {
		callbackHandler.HandleLLMStart(ctx, []string{prompt})
	}

	// 构建请求
	request := kimiclient.ChatRequest{
		Model: o.config.Model,
		Messages: []kimiclient.ChatMessage{
			{
				Role:    RoleUser,
				Content: prompt,
			},
		},
		Temperature: o.config.Temperature,
		TopP:        o.config.TopP,
		MaxTokens:   o.config.MaxTokens,
		Stream:      true,
	}

	// 发送请求
	chunkChan, errChan := o.client.CreateChatStream(ctx, &request)

	// 创建流式响应
	return &streamingResponse{
		ctx:              ctx,
		chunkChan:        chunkChan,
		errChan:          errChan,
		callbacksHandler: callbackHandler,
		text:             strings.Builder{},
	}, nil
}

// StreamingGenerateContent 执行流式内容生成，支持多模态输入和工具调用
func (o *LLM) StreamingGenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (StreamingResponse, error) {
	// 处理调用选项
	llmOptions := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmOptions)
	}

	// 处理回调
	callbackHandler := o.config.CallbacksHandler

	if callbackHandler != nil {
		callbackHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	// 转换消息格式
	kimiMessages, err := convertToKimiMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("转换消息格式失败: %w", err)
	}

	// 构建请求参数
	request := kimiclient.ChatRequest{
		Model:       o.config.Model,
		Messages:    kimiMessages,
		Temperature: o.config.Temperature,
		TopP:        o.config.TopP,
		MaxTokens:   o.config.MaxTokens,
		Stream:      true,
	}

	// 处理工具调用
	if llmOptions.Tools != nil && len(llmOptions.Tools) > 0 {
		tools, err := convertTools(llmOptions.Tools)
		if err != nil {
			return nil, fmt.Errorf("转换工具失败: %w", err)
		}
		request.Tools = tools

		// 处理工具选择
		if llmOptions.ToolChoice != nil {
			toolChoice, err := convertToolChoice(llmOptions.ToolChoice)
			if err != nil {
				return nil, fmt.Errorf("转换工具选择失败: %w", err)
			}
			request.ToolChoice = toolChoice
		}
	}

	// 发送请求
	chunkChan, errChan := o.client.CreateChatStream(ctx, &request)

	// 创建流式响应
	return &streamingContentResponse{
		ctx:              ctx,
		chunkChan:        chunkChan,
		errChan:          errChan,
		callbacksHandler: callbackHandler,
		text:             strings.Builder{},
		toolCalls:        []llms.ToolCall{},
		currentChoice:    &llms.ContentChoice{},
	}, nil
}
