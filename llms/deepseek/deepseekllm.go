package deepseek

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
	// ErrUnexpectedResponseLength is returned when the response length is unexpected.
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	// ErrInvalidContentType is returned when the content type is invalid.
	ErrInvalidContentType = errors.New("invalid content type")
	// ErrUnsupportedMessageType is returned when the message type is unsupported.
	ErrUnsupportedMessageType = errors.New("unsupported message type")
	// ErrUnsupportedContentType is returned when the content type is unsupported.
	ErrUnsupportedContentType = errors.New("unsupported content type")
)

const (
	DefaultModel  = "deepseek-chat"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

// LLM is a DeepSeek large language model.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *deepseekclient.Client
}

var _ llms.Model = (*LLM)(nil)

// New creates a new DeepSeek LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("deepseek: failed to create client: %w", err)
	}
	return &LLM{
		client: c,
	}, nil
}

func newClient(opts ...Option) (*deepseekclient.Client, error) {
	options := &Options{
		APIKey:     os.Getenv(deepseekclient.TokenEnvVarName),
		BaseURL:    os.Getenv(deepseekclient.BaseURLEnvVarName),
		Model:      os.Getenv(deepseekclient.ModelEnvVarName),
		HTTPClient: http.DefaultClient,
	}

	if options.BaseURL == "" {
		options.BaseURL = deepseekclient.DefaultBaseURL
	}
	if options.Model == "" {
		options.Model = DefaultModel
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.APIKey == "" {
		return nil, ErrMissingAPIKey
	}

	return deepseekclient.New(
		options.APIKey,
		options.BaseURL,
		options.Model,
		options.HTTPClient,
	)
}

// GetModels 返回DeepSeek支持的模型列表
func (o *LLM) GetModels() []string {
	return []string{
		"deepseek-chat",        // DeepSeek聊天模型
		"deepseek-coder",       // DeepSeek代码模型
		"deepseek-reasoner",    // DeepSeek推理模型（支持思维链）
		"deepseek-vision",      // DeepSeek视觉模型（多模态）
	}
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	return generateMessagesContent(ctx, o, messages, opts)
}

func generateMessagesContent(ctx context.Context, o *LLM, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
	deepseekMessages, err := convertToDeepSeekMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("deepseek: failed to process messages: %w", err)
	}

	tools := convertTools(opts.Tools)
	if opts.Model == "" {
		opts.Model = o.client.Model
	}

	request := deepseekclient.ChatRequest{
		Model:            opts.Model,
		Messages:         deepseekMessages,
		MaxTokens:        opts.MaxTokens,
		Temperature:      opts.Temperature,
		TopP:             opts.TopP,
		N:                opts.N,
		Stop:             opts.StopWords,
		FrequencyPenalty: opts.FrequencyPenalty,
		PresencePenalty:  opts.PresencePenalty,
		Stream:           opts.StreamingFunc != nil,
		StreamingFunc:    opts.StreamingFunc,
		Tools:            tools,
		ToolChoice:       convertToolChoice(opts.ToolChoice),
	}

	// 处理JSON模式
	if opts.JSONMode {
		request.JSONMode = true
	}

	// 处理种子
	if opts.Seed != 0 {
		request.Seed = opts.Seed
	}

	// 发送请求
	resp, err := o.client.CreateChat(ctx, request)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("deepseek: failed to create chat: %w", err)
	}

	if len(resp.Choices) == 0 {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, ErrEmptyResponse)
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

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, contentResponse)
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
				return nil, fmt.Errorf("deepseek: unsupported content part type: %T", p)
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
		return RoleSystem, nil
	case llms.ChatMessageTypeHuman:
		return RoleUser, nil
	case llms.ChatMessageTypeAI:
		return RoleAssistant, nil
	case llms.ChatMessageTypeTool:
		return RoleTool, nil
	case llms.ChatMessageTypeGeneric:
		return RoleUser, nil
	case llms.ChatMessageTypeFunction:
		return RoleTool, nil
	default:
		return "", fmt.Errorf("deepseek: %w: %v", ErrUnsupportedMessageType, messageType)
	}
}

// convertTools converts langchaingo tools to DeepSeek tools.
func convertTools(tools []llms.Tool) []deepseekclient.Tool {
	if len(tools) == 0 {
		return nil
	}

	toolReq := make([]deepseekclient.Tool, len(tools))
	for i, tool := range tools {
		toolReq[i] = deepseekclient.Tool{
			Type: "function",
			Function: &deepseekclient.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}
	return toolReq
}

// convertToolChoice converts langchaingo tool choice to DeepSeek tool choice.
func convertToolChoice(toolChoice any) any {
	if toolChoice == nil {
		return nil
	}

	switch tc := toolChoice.(type) {
	case string:
		if tc == "auto" || tc == "none" {
			return tc
		}
		return nil
	case llms.ToolChoice:
		return deepseekclient.ToolChoice{
			Type: "function",
			Function: &deepseekclient.FunctionReference{
				Name: tc.Function.Name,
			},
		}
	default:
		return nil
	}
}
