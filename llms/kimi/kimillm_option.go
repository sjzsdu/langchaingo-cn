package kimi

import (
	"net/http"
	"os"

	"github.com/tmc/langchaingo/callbacks"
)

// options 包含创建Kimi LLM客户端的选项
type options struct {
	// apiKey 是Kimi API密钥
	apiKey string

	// model 是要使用的模型名称
	model string

	// baseURL 是API的基础URL
	baseURL string

	// httpClient 是自定义的HTTP客户端
	httpClient *http.Client

	// callbacksHandler 是回调处理器
	callbacksHandler callbacks.Handler

	// temperature 控制随机性，值越高回复越随机
	temperature float64

	// topP 控制词汇选择的多样性
	topP float64

	// maxTokens 是生成的最大令牌数
	maxTokens int
}

// Option 是配置Kimi LLM客户端的函数类型
type Option func(*options)

// defaultOptions 返回默认选项
func defaultOptions() *options {
	return &options{
		apiKey:       os.Getenv("KIMI_API_KEY"),
		model:        os.Getenv("KIMI_MODEL"),
		baseURL:      "https://api.moonshot.cn/v1",
		temperature:  0.7,
		topP:         1.0,
		maxTokens:    2048,
	}
}

// WithToken 设置API密钥
func WithToken(token string) Option {
	return func(o *options) {
		o.apiKey = token
	}
}

// WithModel 设置模型名称
func WithModel(model string) Option {
	return func(o *options) {
		o.model = model
	}
}

// WithBaseURL 设置API的基础URL
func WithBaseURL(baseURL string) Option {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

// WithHTTPClient 设置自定义的HTTP客户端
func WithHTTPClient(client *http.Client) Option {
	return func(o *options) {
		o.httpClient = client
	}
}

// WithCallbacksHandler 设置回调处理器
func WithCallbacksHandler(handler callbacks.Handler) Option {
	return func(o *options) {
		o.callbacksHandler = handler
	}
}

// WithTemperature 设置温度参数
func WithTemperature(temperature float64) Option {
	return func(o *options) {
		o.temperature = temperature
	}
}

// WithTopP 设置topP参数
func WithTopP(topP float64) Option {
	return func(o *options) {
		o.topP = topP
	}
}

// WithMaxTokens 设置最大令牌数
func WithMaxTokens(maxTokens int) Option {
	return func(o *options) {
		o.maxTokens = maxTokens
	}
}