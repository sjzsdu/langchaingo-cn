package zhipu

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	// 环境变量名
	TokenEnvVarName = "ZHIPU_API_KEY" //nolint:gosec
	ModelEnvVarName = "ZHIPU_MODEL"   //nolint:gosec
	// Embedding 环境变量
	EmbeddingModelEnvVarName = "ZHIPU_EMBEDDING_MODEL" //nolint:gosec

	// OpenAI兼容模式基础URL
	OpenAICompatibleBaseURL = "https://open.bigmodel.cn/api/paas/v4/"
	// 默认模型
	DefaultModel = "glm-4"
	// 默认 Embedding 模型
	DefaultEmbeddingModel = "embedding-2"
)

const (
	// ModelGLM4 是智谱GLM-4模型
	ModelGLM4 = "glm-4"

	// ModelGLM4V 是智谱GLM-4V视觉模型
	ModelGLM4V = "glm-4v"

	// ModelGLM4Air 是智谱GLM-4-Air轻量级模型
	ModelGLM4Air = "glm-4-air"

	// ModelGLM4AirX 是智谱GLM-4-AirX模型
	ModelGLM4AirX = "glm-4-airx"

	// ModelGLM4Flash 是智谱GLM-4-Flash快速模型
	ModelGLM4Flash = "glm-4-flash"

	// ModelGLM3Turbo 是智谱GLM-3-Turbo模型
	ModelGLM3Turbo = "glm-3-turbo"

	// ModelCharGLM3 是智谱CharGLM-3角色扮演模型
	ModelCharGLM3 = "charglm-3"

	// ModelCogView3 是智谱CogView-3图像生成模型
	ModelCogView3 = "cogview-3"
)

// LLM 是智谱AI大语言模型的实现
type LLM struct {
	*openai.LLM // 匿名嵌入OpenAI LLM，自动继承其所有方法
}

// Option 是LLM的配置选项函数类型
type Option func(*options)

// options 是LLM的配置选项
type options struct {
	apiKey         string
	baseURL        string
	model          string
	embeddingModel string
}

// WithAPIKey 设置API密钥
func WithAPIKey(apiKey string) Option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

// WithBaseURL 设置基础URL
func WithBaseURL(baseURL string) Option {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

// WithModel 设置模型
func WithModel(model string) Option {
	return func(o *options) {
		o.model = model
	}
}

// WithEmbeddingModel 设置embedding模型
func WithEmbeddingModel(model string) Option {
	return func(o *options) {
		o.embeddingModel = model
	}
}

// defaultOptions 返回默认选项
func defaultOptions() options {
	return options{
		apiKey:         os.Getenv(TokenEnvVarName),
		model:          getEnvOrDefault(ModelEnvVarName, DefaultModel),
		embeddingModel: getEnvOrDefault(EmbeddingModelEnvVarName, DefaultEmbeddingModel),
	}
}

// getEnvOrDefault 获取环境变量值，如果不存在则返回默认值
func getEnvOrDefault(envVar, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	return value
}

// New 创建一个新的智谱AI LLM实例
func New(opts ...Option) (*LLM, error) {
	options := defaultOptions()

	// 应用选项
	for _, opt := range opts {
		opt(&options)
	}

	// 验证API密钥
	if options.apiKey == "" {
		return nil, errors.New("API密钥不能为空，请设置ZHIPU_API_KEY环境变量或使用WithAPIKey选项")
	}

	// 创建OpenAI客户端
	openaiOpts := []openai.Option{
		openai.WithToken(options.apiKey),
		openai.WithModel(options.model),
		openai.WithBaseURL(OpenAICompatibleBaseURL),
		openai.WithEmbeddingModel(options.embeddingModel),
	}

	openaiLLM, err := openai.New(openaiOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建OpenAI客户端失败: %w", err)
	}

	return &LLM{LLM: openaiLLM}, nil
}

// GenerateContent 重写生成内容方法，自动处理system消息转换
func (z *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// 转换system消息为user消息，因为智谱AI不支持system角色
	convertedMessages := make([]llms.MessageContent, 0, len(messages))
	
	for _, msg := range messages {
		if msg.Role == llms.ChatMessageTypeSystem {
			// 将system消息转换为user消息，并在内容前添加提示
			convertedMsg := llms.MessageContent{
				Role: llms.ChatMessageTypeHuman,
				Parts: make([]llms.ContentPart, 0, len(msg.Parts)),
			}
			
			// 处理每个部分
			for i, part := range msg.Parts {
				switch p := part.(type) {
				case llms.TextContent:
					// 在第一个文本内容前添加角色说明
					if i == 0 {
						convertedMsg.Parts = append(convertedMsg.Parts, llms.TextContent{
							Text: "请按以下角色要求回答：" + p.Text,
						})
					} else {
						convertedMsg.Parts = append(convertedMsg.Parts, p)
					}
				default:
					convertedMsg.Parts = append(convertedMsg.Parts, p)
				}
			}
			convertedMessages = append(convertedMessages, convertedMsg)
		} else {
			convertedMessages = append(convertedMessages, msg)
		}
	}
	
	// 调用父类方法
	return z.LLM.GenerateContent(ctx, convertedMessages, options...)
}
