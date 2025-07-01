package deepseek

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek/internal/deepseekclient"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

var (
	// ErrEmptyResponse is returned when the API returns an empty response.
	ErrEmptyResponse = errors.New("empty response")
	// ErrMissingAPIKey is returned when the API key is missing.
	ErrMissingAPIKey = errors.New("missing API key")
	// ErrMissingModel is returned when the model is missing.
	ErrMissingModel = errors.New("missing model")
)

// LLM is a DeepSeek large language model.
type LLM struct {
	client           *deepseekclient.Client
	defaultModel     string
	callbacksHandler callbacks.Handler
}

// New creates a new DeepSeek LLM.
func New(opts ...Option) (*LLM, error) {
	options := DefaultOptions()

	for _, opt := range opts {
		opt(options)
	}

	// 如果没有提供 APIKey，尝试从环境变量获取
	if options.APIKey == "" {
		options.APIKey = os.Getenv(deepseekclient.TokenEnvVarName)
		if options.APIKey == "" {
			return nil, ErrMissingAPIKey
		}
	}

	// 如果没有提供 Model，尝试从环境变量获取
	if options.Model == "" {
		// 从 deepseekllm_option.go 中导入的常量
		options.Model = os.Getenv(deepseekclient.ModelEnvVarName)
		if options.Model == "" {
			return nil, ErrMissingModel
		}
	}

	client, err := deepseekclient.New(
		options.APIKey,
		options.BaseURL,
		options.HTTPClient,
	)
	if err != nil {
		return nil, err
	}

	return &LLM{
		client:           client,
		defaultModel:     options.Model,
		callbacksHandler: options.CallbacksHandler,
	}, nil
}

// Call calls the DeepSeek API with the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	result, err := o.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}

	if len(result) == 0 {
		return "", ErrEmptyResponse
	}

	return result[0], nil
}

// Generate generates completions for multiple prompts.
func (o *LLM) Generate(
	ctx context.Context,
	prompts []string,
	options ...llms.CallOption,
) ([]string, error) {
	callOpts := llms.CallOptions{}
	for _, opt := range options {
		opt(&callOpts)
	}

	model := o.defaultModel
	if callOpts.Model != "" {
		model = callOpts.Model
	}

	results := make([]string, 0, len(prompts))

	for _, prompt := range prompts {
		messages := []deepseekclient.ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		// 构建请求参数
		request := deepseekclient.ChatRequest{
			Model:            model,
			Messages:         messages,
			MaxTokens:        callOpts.MaxTokens,
			Temperature:      callOpts.Temperature,
			TopP:             callOpts.TopP,
			N:                callOpts.N,
			Stop:             callOpts.StopWords,
			FrequencyPenalty: callOpts.FrequencyPenalty,
			PresencePenalty:  callOpts.PresencePenalty,
			Stream:           callOpts.StreamingFunc != nil,
			StreamingFunc:    callOpts.StreamingFunc,
		}

		// 处理JSON模式
		if callOpts.JSONMode {
			request.JSONMode = true
		}

		// 处理种子
		if callOpts.Seed != 0 {
			request.Seed = callOpts.Seed
		}

		// 发送请求
		resp, err := o.client.CreateChat(ctx, request)
		if err != nil {
			if o.callbacksHandler != nil {
				o.callbacksHandler.HandleLLMError(ctx, err)
			}
			return nil, err
		}

		if len(resp.Choices) == 0 {
			if o.callbacksHandler != nil {
				o.callbacksHandler.HandleLLMError(ctx, ErrEmptyResponse)
			}
			return nil, ErrEmptyResponse
		}

		// 处理生成结果
		for _, choice := range resp.Choices {
			results = append(results, choice.Message.Content)
		}
	}

	return results, nil
}

// GenerateContent generates content based on the provided parts.
func (o *LLM) GenerateContent(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	if o.callbacksHandler != nil {
		o.callbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	callOpts := llms.CallOptions{}
	for _, opt := range options {
		opt(&callOpts)
	}

	model := o.defaultModel
	if callOpts.Model != "" {
		model = callOpts.Model
	}

	deepseekMessages, err := convertToDeepSeekMessages(messages)
	if err != nil {
		if o.callbacksHandler != nil {
			o.callbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	// 构建请求参数
	request := deepseekclient.ChatRequest{
		Model:            model,
		Messages:         deepseekMessages,
		MaxTokens:        callOpts.MaxTokens,
		Temperature:      callOpts.Temperature,
		TopP:             callOpts.TopP,
		N:                callOpts.N,
		Stop:             callOpts.StopWords,
		FrequencyPenalty: callOpts.FrequencyPenalty,
		PresencePenalty:  callOpts.PresencePenalty,
		Stream:           callOpts.StreamingFunc != nil,
		StreamingFunc:    callOpts.StreamingFunc,
		Tools:            convertTools(callOpts.Tools),
		ToolChoice:       convertToolChoice(callOpts.ToolChoice),
		JSONMode:         callOpts.JSONMode,
	}

	// 处理种子
	if callOpts.Seed != 0 {
		request.Seed = callOpts.Seed
	}

	// 发送请求
	resp, err := o.client.CreateChat(ctx, request)
	if err != nil {
		if o.callbacksHandler != nil {
			o.callbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	if len(resp.Choices) == 0 {
		if o.callbacksHandler != nil {
			o.callbacksHandler.HandleLLMError(ctx, ErrEmptyResponse)
		}
		return nil, ErrEmptyResponse
	}

	// 构建响应
	contentResponse := &llms.ContentResponse{
		Choices: make([]*llms.ContentChoice, len(resp.Choices)),
	}

	// 处理选择
	for i, choice := range resp.Choices {
		contentChoice := &llms.ContentChoice{
			StopReason: choice.FinishReason,
		}

		// 处理内容
		if choice.Message.Content != "" {
			contentChoice.Content = choice.Message.Content
		}

		// 处理推理内容
		if choice.Message.ReasoningContent != "" {
			contentChoice.ReasoningContent = choice.Message.ReasoningContent
		}

		// 处理工具调用
		if len(choice.Message.ToolCalls) > 0 {
			// 设置第一个工具调用为FuncCall
			toolCall := choice.Message.ToolCalls[0]
			contentChoice.FuncCall = &llms.FunctionCall{
				Name:      toolCall.Function.Name,
				Arguments: toolCall.Function.Arguments,
			}

			// 设置所有工具调用
			contentChoice.ToolCalls = make([]llms.ToolCall, len(choice.Message.ToolCalls))
			for j, tc := range choice.Message.ToolCalls {
				contentChoice.ToolCalls[j] = llms.ToolCall{
					ID:   tc.ID,
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		contentResponse.Choices[i] = contentChoice
	}

	if o.callbacksHandler != nil {
		o.callbacksHandler.HandleLLMGenerateContentEnd(ctx, contentResponse)
	}

	return contentResponse, nil
}

// convertToDeepSeekMessages converts langchaingo message content to DeepSeek messages.
func convertToDeepSeekMessages(messages []llms.MessageContent) ([]deepseekclient.ChatMessage, error) {
	deepseekMessages := make([]deepseekclient.ChatMessage, 0, len(messages))

	for _, message := range messages {
		role, err := convertMessageType(message.Role)
		if err != nil {
			return nil, err
		}

		// 处理纯文本内容
		if len(message.Parts) == 1 {
			if textContent, ok := message.Parts[0].(llms.TextContent); ok {
				deepseekMessages = append(deepseekMessages, deepseekclient.ChatMessage{
					Role:    role,
					Content: textContent.Text,
				})
				continue
			}
		}

		// 处理多模态内容
		contentParts := make([]deepseekclient.ContentPart, 0, len(message.Parts))
		// 检查是否有工具调用
		var toolCalls []deepseekclient.ToolCall
		for _, part := range message.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				contentParts = append(contentParts, deepseekclient.ContentPart{
					Type: "text",
					Text: p.Text,
				})
			case llms.ImageURLContent:
				contentParts = append(contentParts, deepseekclient.ContentPart{
					Type: "image_url",
					ImageURL: &deepseekclient.ImageURL{
						URL:    p.URL,
						Detail: p.Detail,
					},
				})
			case llms.ToolCall:
				// 收集工具调用
				if role == "assistant" {
					toolCalls = append(toolCalls, deepseekclient.ToolCall{
						ID:   p.ID,
						Type: p.Type,
						Function: &deepseekclient.FunctionCall{
							Name:      p.FunctionCall.Name,
							Arguments: p.FunctionCall.Arguments,
						},
					})
				}
			case llms.ToolCallResponse:
				// 工具调用响应作为单独的消息处理
				deepseekMessages = append(deepseekMessages, deepseekclient.ChatMessage{
					Role:       "tool",
					Content:    p.Content,
					ToolCallID: p.ToolCallID,
					Name:       p.Name,
				})
				// 跳过将此部分添加到当前消息的内容部分
				continue
			default:
				return nil, fmt.Errorf("unsupported content part type: %T", p)
			}
		}

		// 如果有多模态内容，创建带有内容部分的消息
		if len(contentParts) > 0 {
			chatMessage := deepseekclient.ChatMessage{
				Role:         role,
				ContentParts: contentParts,
			}

			deepseekMessages = append(deepseekMessages, chatMessage)
		} else if role == "assistant" && len(toolCalls) > 0 {
			// 如果没有内容部分但有工具调用，创建一个带有工具调用的消息
			// DeepSeek API 要求 assistant 消息必须设置 content 和 tool_calls 字段

			chatMessage := deepseekclient.ChatMessage{
				Role:      role,
				Content:   " ",       // 设置为空格字符串而不是空字符串，确保 content 字段不为空
				ToolCalls: toolCalls, // 添加工具调用到消息中
			}

			deepseekMessages = append(deepseekMessages, chatMessage)
		}
	}

	return deepseekMessages, nil
}

// convertMessageType converts langchaingo message type to DeepSeek role.
func convertMessageType(messageType llms.ChatMessageType) (string, error) {
	switch messageType {
	case llms.ChatMessageTypeSystem:
		return "system", nil
	case llms.ChatMessageTypeAI:
		return "assistant", nil
	case llms.ChatMessageTypeHuman:
		return "user", nil
	case llms.ChatMessageTypeGeneric:
		return "user", nil
	case llms.ChatMessageTypeFunction:
		return "function", nil
	case llms.ChatMessageTypeTool:
		return "tool", nil
	default:
		return "", fmt.Errorf("unsupported message type: %v", messageType)
	}
}

// convertTools converts langchaingo tools to DeepSeek tools.
func convertTools(tools []llms.Tool) []deepseekclient.Tool {
	if len(tools) == 0 {
		return nil
	}

	deepseekTools := make([]deepseekclient.Tool, len(tools))
	for i, tool := range tools {
		deepseekTools[i] = deepseekclient.Tool{
			Type: tool.Type,
		}

		if tool.Function != nil {
			deepseekTools[i].Function = &deepseekclient.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			}
		}
	}

	return deepseekTools
}

// convertToolChoice converts langchaingo tool choice to DeepSeek tool choice.
func convertToolChoice(toolChoice any) any {
	if toolChoice == nil {
		return nil
	}

	switch tc := toolChoice.(type) {
	case string:
		return tc
	case llms.ToolChoice:
		deepseekToolChoice := deepseekclient.ToolChoice{
			Type: tc.Type,
		}

		if tc.Function != nil {
			deepseekToolChoice.Function = &deepseekclient.FunctionReference{
				Name: tc.Function.Name,
			}
		}

		return deepseekToolChoice
	default:
		return toolChoice
	}
}
