package deepseek

import (
	"net/http"

	"github.com/tmc/langchaingo/callbacks"
)

// Option is a function that configures a DeepSeek LLM.
type Option func(*Options)

// Options is the configuration for a DeepSeek LLM.
type Options struct {
	// APIKey is the API key for the DeepSeek API.
	APIKey string
	// Model is the model to use.
	Model string
	// BaseURL is the base URL for the DeepSeek API.
	BaseURL string
	// HTTPClient is the HTTP client to use.
	HTTPClient *http.Client
	// CallbacksHandler is the callbacks handler to use.
	CallbacksHandler callbacks.Handler
}

// DefaultOptions returns the default options for the DeepSeek LLM.
func DefaultOptions() *Options {
	return &Options{
		BaseURL:    "https://api.deepseek.com",
		Model:      "deepseek-chat", // 默认使用 DeepSeek-V3 模型
		HTTPClient: http.DefaultClient,
	}
}

// WithAPIKey sets the API key for the DeepSeek API.
func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

// WithModel sets the model to use.
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = model
	}
}

// WithBaseURL sets the base URL for the DeepSeek API.
func WithBaseURL(baseURL string) Option {
	return func(o *Options) {
		o.BaseURL = baseURL
	}
}

// WithHTTPClient sets the HTTP client to use.
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		o.HTTPClient = client
	}
}

// WithCallbacksHandler sets the callbacks handler to use.
func WithCallbacksHandler(callbacksHandler callbacks.Handler) Option {
	return func(o *Options) {
		o.CallbacksHandler = callbacksHandler
	}
}
