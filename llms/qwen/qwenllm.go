// Package qwen 提供了阿里云通义千问大语言模型的Go语言客户端实现
package qwen

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sjzsdu/langchaingo-cn/llms/qwen/internal/qwenclient"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

// StreamingResponse 定义了流式响应的接口
type StreamingResponse interface {
	// GetChunk 获取下一个内容块
	GetChunk() (string, error)
}

const (
	// ModelQWenTurbo 是通义千问Turbo模型
	ModelQWenTurbo = "qwen-turbo"

	// ModelQWenPlus 是通义千问Plus模型
	ModelQWenPlus = "qwen-plus"

	// ModelQWenMax 是通义千问Max模型
	ModelQWenMax = "qwen-max"

	// ModelQWenVLPlus 是通义千问视觉Plus模型
	ModelQWenVLPlus = "qwen-vl-plus"

	// ModelQWenVLMax 是通义千问视觉Max模型
	ModelQWenVLMax = "qwen-vl-max"

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
	ErrMissingAPIKey = errors.New("缺少API密钥，请通过参数或环境变量 QWEN_API_KEY 设置")

	// ErrMissingModel 表示缺少模型名称
	ErrMissingModel = errors.New("缺少模型名称，请通过参数或环境变量 QWEN_MODEL 设置")

	// ErrRequestFailed 表示请求失败
	ErrRequestFailed = errors.New("请求失败")
)

// LLM 是通义千问大语言模型的客户端
type LLM struct {
	// CallbacksHandler 是回调处理器
	CallbacksHandler callbacks.Handler

	// client 是通义千问API客户端
	client *qwenclient.Client

	// QWenModel 是要使用的模型名称
	QWenModel string

	// Temperature 控制随机性，值越高回复越随机
	Temperature float64

	// TopP 控制词汇选择的多样性
	TopP float64

	// TopK 控制每一步考虑的词汇数量
	TopK int

	// MaxTokens 是生成的最大令牌数
	MaxTokens int

	// UseOpenAICompatible 表示是否使用OpenAI兼容模式
	UseOpenAICompatible bool
}

// New 创建一个新的通义千问LLM客户端
func New(opts ...Option) (*LLM, error) {
	options := defaultOptions()

	for _, opt := range opts {
		opt(options)
	}

	if options.apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	if options.model == "" {
		return nil, ErrMissingModel
	}

	// 创建通义千问API客户端
	clientOpts := []qwenclient.ClientOption{
		qwenclient.WithOpenAICompatible(options.useOpenAICompatible),
	}

	if options.baseURL != "" {
		clientOpts = append(clientOpts, qwenclient.WithBaseURL(options.baseURL))
	}

	if options.httpClient != nil {
		clientOpts = append(clientOpts, qwenclient.WithHTTPClient(options.httpClient))
	}

	client, err := qwenclient.NewClient(options.apiKey, clientOpts...)
	if err != nil {
		// 提供更详细的错误信息
		return nil, fmt.Errorf("创建通义千问客户端失败: %w (请确保提供了有效的API密钥)", err)
	}

	return &LLM{
		CallbacksHandler:    options.callbacksHandler,
		client:              client,
		QWenModel:           options.model,
		Temperature:         options.temperature,
		TopP:                options.topP,
		TopK:                options.topK,
		MaxTokens:           options.maxTokens,
		UseOpenAICompatible: options.useOpenAICompatible,
	}, nil
}

// Call 调用通义千问API生成文本
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	llmOptions := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmOptions)
	}

	// 处理回调
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, []string{prompt})
	}

	// 构建请求
	request := qwenclient.ChatRequest{
		Model: o.QWenModel,
		Input: qwenclient.ChatRequestInput{
			Prompt: prompt,
		},
		Parameters: qwenclient.ChatRequestParameters{
			Temperature: o.Temperature,
			TopP:        o.TopP,
			TopK:        o.TopK,
			MaxTokens:   o.MaxTokens,
		},
	}

	// 处理JSON模式
	if llmOptions.JSONMode {
		request.Parameters.ResultFormat = "json"
	}

	// 处理种子
	if llmOptions.Seed != 0 {
		request.Parameters.Seed = int64(llmOptions.Seed)
	}

	// 发送请求
	response, err := o.client.CreateChat(ctx, request)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return "", fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	// 处理响应
	if response == nil || response.Output.Text == "" {
		return "", ErrEmptyResponse
	}

	// 处理回调
	if o.CallbacksHandler != nil {
		// 创建一个ContentResponse用于回调
		contentResponse := &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					Content: response.Output.Text,
				},
			},
		}
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, contentResponse)
	}

	return response.Output.Text, nil
}

// Generate 生成文本
func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]string, error) {
	results := make([]string, 0, len(prompts))

	for _, prompt := range prompts {
		response, err := o.Call(ctx, prompt, options...)
		if err != nil {
			return nil, err
		}

		results = append(results, response)
	}

	return results, nil
}

// GenerateContent 生成内容，支持多模态输入和工具调用
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// 处理回调
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	// 解析选项
	llmOptions := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmOptions)
	}

	// 转换消息格式
	qwenMessages, err := convertToQWenMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("转换消息格式失败: %w", err)
	}

	// 构建请求参数
	request := qwenclient.ChatRequest{
		Model: o.QWenModel,
		Input: qwenclient.ChatRequestInput{
			Messages: qwenMessages,
		},
		Parameters: qwenclient.ChatRequestParameters{
			Temperature: o.Temperature,
			TopP:        o.TopP,
			TopK:        o.TopK,
			MaxTokens:   o.MaxTokens,
		},
	}

	// 处理工具调用
	if llmOptions.Tools != nil && len(llmOptions.Tools) > 0 {
		// 将[]llms.Tool转换为[]interface{}
		toolsInterface := make([]interface{}, len(llmOptions.Tools))
		for i, tool := range llmOptions.Tools {
			toolsInterface[i] = tool
		}

		tools, err := convertTools(toolsInterface)
		if err != nil {
			return nil, fmt.Errorf("转换工具失败: %w", err)
		}
		request.Parameters.Tools = tools

		// 处理工具选择
		if llmOptions.ToolChoice != nil {
			toolChoice, err := convertToolChoice(llmOptions.ToolChoice)
			if err != nil {
				return nil, fmt.Errorf("转换工具选择失败: %w", err)
			}
			request.Parameters.ToolChoice = toolChoice
		}
	}

	// 发送请求
	response, err := o.client.CreateChat(ctx, request)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	// 处理响应
	if response == nil {
		return nil, ErrEmptyResponse
	}

	// 创建响应
	contentResponse := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: response.Output.Text,
			},
		},
	}

	// 处理工具调用
	if len(response.Output.ToolCalls) > 0 {
		toolCalls := make([]llms.ToolCall, 0, len(response.Output.ToolCalls))
		for _, tc := range response.Output.ToolCalls {
			toolCalls = append(toolCalls, llms.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				FunctionCall: &llms.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
		contentResponse.Choices[0].ToolCalls = toolCalls

		// 如果有工具调用，设置第一个工具调用到FuncCall字段
		if len(toolCalls) > 0 {
			contentResponse.Choices[0].FuncCall = toolCalls[0].FunctionCall
		}
	}

	// 处理回调
	if o.CallbacksHandler != nil {
		// 直接使用contentResponse作为回调参数
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, contentResponse)
	}

	return contentResponse, nil
}

// convertToQWenMessages 将LangChainGo消息转换为通义千问消息
func convertToQWenMessages(messages []llms.MessageContent) ([]qwenclient.ChatMessage, error) {
	qwenMessages := make([]qwenclient.ChatMessage, 0, len(messages))

	for _, message := range messages {
		role := convertMessageType(string(message.Role))

		// 处理纯文本消息
		if len(message.Parts) == 0 {
			qwenMessages = append(qwenMessages, qwenclient.ChatMessage{
				Role:    role,
				Content: "",
			})
			continue
		}

		// 检查是否只有一个文本部分
		if len(message.Parts) == 1 {
			if textContent, ok := message.Parts[0].(llms.TextContent); ok {
				qwenMessages = append(qwenMessages, qwenclient.ChatMessage{
					Role:    role,
					Content: textContent.Text,
				})
				continue
			}
		}

		// 处理多模态内容
		contentParts := make([]qwenclient.ContentPart, 0, len(message.Parts))
		for _, content := range message.Parts {
			switch c := content.(type) {
			case llms.TextContent:
				contentParts = append(contentParts, qwenclient.ContentPart{
					Type: "text",
					Text: c.Text,
				})
			case llms.ImageURLContent:
				// 处理图像URL内容
				contentParts = append(contentParts, qwenclient.ContentPart{
					Type: "image",
					ImageURL: &qwenclient.ImageURL{
						URL: c.URL,
					},
				})
			case llms.BinaryContent:
				// 处理二进制内容
				base64Data := base64.StdEncoding.EncodeToString(c.Data)
				imageURL := "data:" + c.MIMEType + ";base64," + base64Data
				contentParts = append(contentParts, qwenclient.ContentPart{
					Type: "image",
					ImageURL: &qwenclient.ImageURL{
						URL: imageURL,
					},
				})
			default:
				return nil, fmt.Errorf("不支持的内容类型: %T", c)
			}
		}

		qwenMessages = append(qwenMessages, qwenclient.ChatMessage{
			Role:         role,
			ContentParts: contentParts,
		})
	}

	return qwenMessages, nil
}

// convertMessageType 将LangChainGo消息类型转换为通义千问角色
func convertMessageType(messageType string) string {
	switch messageType {
	case "human":
		return RoleUser
	case "ai":
		return RoleAssistant
	case "system":
		return RoleSystem
	case "tool":
		return RoleTool
	default:
		return RoleUser
	}
}

// convertTools 将LangChainGo工具转换为通义千问工具
func convertTools(tools []interface{}) ([]qwenclient.Tool, error) {
	qwenTools := make([]qwenclient.Tool, 0, len(tools))

	// 定义工具结构体
	type toolStruct struct {
		Type     string
		Function *struct {
			Name        string
			Description string
			Parameters  interface{}
			Required    []string
		}
	}

	for _, toolInterface := range tools {
		// 将接口转换为JSON，然后解析为结构体
		data, err := json.Marshal(toolInterface)
		if err != nil {
			return nil, fmt.Errorf("序列化工具失败: %w", err)
		}

		var tool toolStruct
		if err := json.Unmarshal(data, &tool); err != nil {
			return nil, fmt.Errorf("解析工具失败: %w", err)
		}

		// 目前只支持函数工具
		if tool.Type != "function" || tool.Function == nil {
			continue
		}

		// 解析函数参数
		var params map[string]interface{}
		if tool.Function.Parameters != nil {
			// 如果参数是字符串，尝试解析JSON
			if paramsStr, ok := tool.Function.Parameters.(string); ok && paramsStr != "" {
				if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
					return nil, fmt.Errorf("解析函数参数失败: %w", err)
				}
			} else {
				// 如果参数是对象，直接使用
				paramsData, err := json.Marshal(tool.Function.Parameters)
				if err != nil {
					return nil, fmt.Errorf("序列化函数参数失败: %w", err)
				}
				if err := json.Unmarshal(paramsData, &params); err != nil {
					return nil, fmt.Errorf("解析函数参数失败: %w", err)
				}
			}
		} else {
			params = make(map[string]interface{})
		}

		qwenTools = append(qwenTools, qwenclient.Tool{
			Type: "function",
			Function: qwenclient.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  params,
				Required:    tool.Function.Required,
			},
		})
	}

	return qwenTools, nil
}

// convertToolChoice 将LangChainGo工具选择转换为通义千问工具选择
func convertToolChoice(toolChoice interface{}) (*qwenclient.ToolChoice, error) {
	if toolChoice == nil {
		return nil, nil
	}

	// 处理字符串类型的工具选择
	if tc, ok := toolChoice.(string); ok {
		if tc == "auto" {
			return &qwenclient.ToolChoice{
				Type: "auto",
			}, nil
		}
		return nil, fmt.Errorf("不支持的工具选择类型: %s", tc)
	}

	// 处理结构体类型的工具选择
	type toolChoiceStruct struct {
		Type     string
		Function *struct {
			Name string
		}
	}

	// 尝试将接口转换为JSON，然后解析为结构体
	data, err := json.Marshal(toolChoice)
	if err != nil {
		return nil, fmt.Errorf("序列化工具选择失败: %w", err)
	}

	var tc toolChoiceStruct
	if err := json.Unmarshal(data, &tc); err != nil {
		return nil, fmt.Errorf("解析工具选择失败: %w", err)
	}

	switch tc.Type {
	case "auto":
		return &qwenclient.ToolChoice{
			Type: "auto",
		}, nil
	case "function":
		if tc.Function == nil {
			return nil, fmt.Errorf("函数工具选择缺少函数定义")
		}
		return &qwenclient.ToolChoice{
			Type: "function",
			Function: &qwenclient.FunctionChoice{
				Name: tc.Function.Name,
			},
		}, nil
	default:
		return nil, fmt.Errorf("不支持的工具选择类型: %s", tc.Type)
	}
}

// streamingResponse 实现了llms.StreamingResponse接口
type streamingResponse struct {
	ctx              context.Context
	chunkChan        <-chan qwenclient.ChatResponseChunk
	errChan          <-chan error
	callbacksHandler callbacks.Handler
	text             strings.Builder
}

// GetChunk 获取下一个文本块
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
				// 创建一个ContentResponse用于回调
				contentResponse := &llms.ContentResponse{
					Choices: []*llms.ContentChoice{
						{
							Content: s.text.String(),
						},
					},
				}
				s.callbacksHandler.HandleLLMGenerateContentEnd(s.ctx, contentResponse)
			}
			return "", io.EOF
		}

		// 处理文本块
		s.text.WriteString(chunk.Output.Text)
		return chunk.Output.Text, nil
	}
}

// streamingContentResponse 实现了StreamingResponse接口
type streamingContentResponse struct {
	ctx              context.Context
	chunkChan        <-chan qwenclient.ChatResponseChunk
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
				// 创建一个ContentResponse用于回调
				contentResponse := &llms.ContentResponse{
					Choices: []*llms.ContentChoice{
						{
							Content:   s.text.String(),
							ToolCalls: s.toolCalls,
						},
					},
				}
				// 如果有工具调用，设置第一个工具调用到FuncCall字段
				if len(s.toolCalls) > 0 {
					contentResponse.Choices[0].FuncCall = s.toolCalls[0].FunctionCall
				}
				s.callbacksHandler.HandleLLMGenerateContentEnd(s.ctx, contentResponse)
			}

			return "", io.EOF
		}

		// 处理文本
		s.text.WriteString(chunk.Output.Text)

		// 处理工具调用
		if len(chunk.Output.ToolCalls) > 0 {
			for _, tc := range chunk.Output.ToolCalls {
				toolCall := llms.ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					FunctionCall: &llms.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}

				// 更新或添加工具调用
				updated := false
				for i, existingTC := range s.toolCalls {
					if existingTC.ID == tc.ID {
						// 更新现有工具调用
						s.toolCalls[i] = toolCall
						updated = true
						break
					}
				}

				if !updated {
					// 添加新工具调用
					s.toolCalls = append(s.toolCalls, toolCall)
				}
			}
		}

		return chunk.Output.Text, nil
	}
}

// StreamingCall 流式调用通义千问API生成文本
func (o *LLM) StreamingCall(ctx context.Context, prompt string, options ...llms.CallOption) (StreamingResponse, error) {
	llmOptions := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmOptions)
	}

	// 处理回调
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, []string{prompt})
	}

	// 构建请求
	request := qwenclient.ChatRequest{
		Model: o.QWenModel,
		Input: qwenclient.ChatRequestInput{
			Prompt: prompt,
		},
		Parameters: qwenclient.ChatRequestParameters{
			Temperature: o.Temperature,
			TopP:        o.TopP,
			TopK:        o.TopK,
			MaxTokens:   o.MaxTokens,
		},
	}

	// 处理JSON模式
	if llmOptions.JSONMode {
		request.Parameters.ResultFormat = "json"
	}

	// 处理种子
	if llmOptions.Seed != 0 {
		request.Parameters.Seed = int64(llmOptions.Seed)
	}

	// 发送流式请求
	chunkChan, errChan, err := o.client.CreateChatStream(ctx, request)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	// 创建流式响应
	return &streamingResponse{
		ctx:              ctx,
		chunkChan:        chunkChan,
		errChan:          errChan,
		callbacksHandler: o.CallbacksHandler,
	}, nil
}

// StreamingGenerateContent 流式生成内容，支持多模态输入和工具调用
func (o *LLM) StreamingGenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (StreamingResponse, error) {
	// 处理回调
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	// 解析选项
	llmOptions := llms.CallOptions{}
	for _, opt := range options {
		opt(&llmOptions)
	}

	// 转换消息格式
	qwenMessages, err := convertToQWenMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("转换消息格式失败: %w", err)
	}

	// 构建请求参数
	request := qwenclient.ChatRequest{
		Model: o.QWenModel,
		Input: qwenclient.ChatRequestInput{
			Messages: qwenMessages,
		},
		Parameters: qwenclient.ChatRequestParameters{
			Temperature: o.Temperature,
			TopP:        o.TopP,
			TopK:        o.TopK,
			MaxTokens:   o.MaxTokens,
		},
	}

	// 处理工具调用
	if llmOptions.Tools != nil && len(llmOptions.Tools) > 0 {
		// 将[]llms.Tool转换为[]interface{}
		toolsInterface := make([]interface{}, len(llmOptions.Tools))
		for i, tool := range llmOptions.Tools {
			toolsInterface[i] = tool
		}

		tools, err := convertTools(toolsInterface)
		if err != nil {
			return nil, fmt.Errorf("转换工具失败: %w", err)
		}
		request.Parameters.Tools = tools

		// 处理工具选择
		if llmOptions.ToolChoice != nil {
			toolChoice, err := convertToolChoice(llmOptions.ToolChoice)
			if err != nil {
				return nil, fmt.Errorf("转换工具选择失败: %w", err)
			}
			request.Parameters.ToolChoice = toolChoice
		}
	}

	// 发送流式请求
	chunkChan, errChan, err := o.client.CreateChatStream(ctx, request)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	// 创建流式内容响应
	return &streamingContentResponse{
		ctx:              ctx,
		chunkChan:        chunkChan,
		errChan:          errChan,
		callbacksHandler: o.CallbacksHandler,
		toolCalls:        make([]llms.ToolCall, 0),
	}, nil
}
