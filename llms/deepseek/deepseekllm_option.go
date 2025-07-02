package deepseek

import (
	"net/http"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek/internal/deepseekclient"
	"github.com/tmc/langchaingo/callbacks"
)

// Options is the configuration for a DeepSeek LLM.
type Options struct {
	APIKey           string
	Model            string
	BaseURL          string
	HTTPClient       deepseekclient.Doer
	CallbacksHandler interface{}
}

// Option is a function that configures a DeepSeek LLM.
type Option func(*Options)

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
