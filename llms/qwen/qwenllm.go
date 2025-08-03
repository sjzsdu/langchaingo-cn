package qwen

import (
	"errors"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms/openai"
)

const (
	// 环境变量名
	TokenEnvVarName = "QWEN_API_KEY" //nolint:gosec
	ModelEnvVarName = "QWEN_MODEL"   //nolint:gosec

	// OpenAI兼容模式基础URL
	OpenAICompatibleBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	// 默认模型
	DefaultModel = "qwen-max"
)

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
)

// LLM 是通义千问大语言模型的实现
type LLM struct {
	*openai.LLM // 匿名嵌入OpenAI LLM，自动继承其所有方法
}

// Option 是LLM的配置选项函数类型
type Option func(*options)

// options 是LLM的配置选项
type options struct {
	apiKey  string
	baseURL string
	model   string
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

// defaultOptions 返回默认选项
func defaultOptions() options {
	return options{
		apiKey: os.Getenv(TokenEnvVarName),
		model:  getEnvOrDefault(ModelEnvVarName, DefaultModel),
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

// New 创建一个新的通义千问LLM实例
func New(opts ...Option) (*LLM, error) {
	options := defaultOptions()

	// 应用选项
	for _, opt := range opts {
		opt(&options)
	}

	// 验证API密钥
	if options.apiKey == "" {
		return nil, errors.New("API密钥不能为空，请设置QWEN_API_KEY环境变量或使用WithAPIKey选项")
	}

	// 创建OpenAI客户端
	openaiOpts := []openai.Option{
		openai.WithToken(options.apiKey),
		openai.WithModel(options.model),
		openai.WithBaseURL(OpenAICompatibleBaseURL),
	}

	openaiLLM, err := openai.New(openaiOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建OpenAI客户端失败: %w", err)
	}

	return &LLM{LLM: openaiLLM}, nil
}
