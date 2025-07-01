// Package qwen 提供了阿里云通义千问大语言模型的Go语言客户端实现
package qwen

import (
	"net/http"
	"os"

	"github.com/tmc/langchaingo/callbacks"
)

// 环境变量常量
const (
	// EnvQWenAPIKey 是通义千问API密钥的环境变量名
	EnvQWenAPIKey = "QWEN_API_KEY"

	// EnvQWenModel 是通义千问模型的环境变量名
	EnvQWenModel = "QWEN_MODEL"

	// EnvQWenBaseURL 是通义千问API基础URL的环境变量名
	EnvQWenBaseURL = "QWEN_BASE_URL"

	// EnvQWenUseOpenAICompatible 是是否使用OpenAI兼容模式的环境变量名
	EnvQWenUseOpenAICompatible = "QWEN_USE_OPENAI_COMPATIBLE"
)

// options 包含通义千问LLM的配置选项
type options struct {
	// apiKey 是通义千问API的密钥
	apiKey string

	// model 是要使用的模型名称
	model string

	// baseURL 是API的基础URL
	baseURL string

	// httpClient 是用于发送HTTP请求的客户端
	httpClient *http.Client

	// callbacksHandler 是回调处理器
	callbacksHandler callbacks.Handler

	// temperature 控制随机性，值越高回复越随机
	temperature float64

	// topP 控制词汇选择的多样性
	topP float64

	// topK 控制每一步考虑的词汇数量
	topK int

	// maxTokens 是生成的最大令牌数
	maxTokens int

	// useOpenAICompatible 表示是否使用OpenAI兼容模式
	useOpenAICompatible bool
}

// Option 是配置通义千问LLM的函数类型
type Option func(*options)

// defaultOptions 返回默认配置选项
func defaultOptions() *options {
	// 从环境变量加载默认配置
	apiKey := os.Getenv(EnvQWenAPIKey)
	model := os.Getenv(EnvQWenModel)
	baseURL := os.Getenv(EnvQWenBaseURL)
	useOpenAICompatible := os.Getenv(EnvQWenUseOpenAICompatible) == "true"

	// 设置默认模型
	if model == "" {
		model = ModelQWenTurbo
	}

	return &options{
		apiKey:              apiKey,
		model:               model,
		baseURL:             baseURL,
		httpClient:          http.DefaultClient,
		callbacksHandler:    nil,
		temperature:         0.7,
		topP:                0.8,
		topK:                50,
		maxTokens:           1024,
		useOpenAICompatible: useOpenAICompatible,
	}
}

// WithAPIKey 设置API密钥
func WithAPIKey(apiKey string) Option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

// WithModel 设置模型名称
func WithModel(model string) Option {
	return func(o *options) {
		o.model = model
	}
}

// WithBaseURL 设置API基础URL
func WithBaseURL(baseURL string) Option {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

// WithHTTPClient 设置HTTP客户端
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *options) {
		o.httpClient = httpClient
	}
}

// WithCallbacksHandler 设置回调处理器
func WithCallbacksHandler(callbacksHandler callbacks.Handler) Option {
	return func(o *options) {
		o.callbacksHandler = callbacksHandler
	}
}

// WithTemperature 设置温度参数
func WithTemperature(temperature float64) Option {
	return func(o *options) {
		o.temperature = temperature
	}
}

// WithTopP 设置TopP参数
func WithTopP(topP float64) Option {
	return func(o *options) {
		o.topP = topP
	}
}

// WithTopK 设置TopK参数
func WithTopK(topK int) Option {
	return func(o *options) {
		o.topK = topK
	}
}

// WithMaxTokens 设置最大令牌数
func WithMaxTokens(maxTokens int) Option {
	return func(o *options) {
		o.maxTokens = maxTokens
	}
}

// WithOpenAICompatible 设置是否使用OpenAI兼容模式
func WithOpenAICompatible(useOpenAICompatible bool) Option {
	return func(o *options) {
		o.useOpenAICompatible = useOpenAICompatible
	}
}
