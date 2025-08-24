package schema

import (
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"

	// 本地LLM包
	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/sjzsdu/langchaingo-cn/llms/kimi"
	"github.com/sjzsdu/langchaingo-cn/llms/qwen"
)

// LLMFactory LLM组件工厂
type LLMFactory struct{}

// NewLLMFactory 创建LLM工厂实例
func NewLLMFactory() *LLMFactory {
	return &LLMFactory{}
}

// Create 根据配置创建LLM实例
func (f *LLMFactory) Create(config *LLMConfig) (llms.Model, error) {
	if config == nil {
		return nil, fmt.Errorf("LLM config is nil")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid LLM config: %w", err)
	}

	// 获取API密钥
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = f.getDefaultAPIKey(config.Type)
	}

	switch config.Type {
	case "openai":
		return f.createOpenAI(config, apiKey)
	case "deepseek":
		return f.createDeepSeek(config, apiKey)
	case "kimi":
		return f.createKimi(config, apiKey)
	case "qwen":
		return f.createQwen(config, apiKey)
	case "anthropic":
		return f.createAnthropic(config, apiKey)
	case "ollama":
		return f.createOllama(config)
	default:
		return nil, fmt.Errorf("unsupported LLM type: %s", config.Type)
	}
}

// createOpenAI 创建OpenAI LLM
func (f *LLMFactory) createOpenAI(config *LLMConfig, apiKey string) (llms.Model, error) {
	var opts []openai.Option

	// 设置API密钥
	if apiKey != "" {
		opts = append(opts, openai.WithToken(apiKey))
	}

	// 设置模型
	if config.Model != "" {
		opts = append(opts, openai.WithModel(config.Model))
	}

	// 设置基础URL
	if config.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(config.BaseURL))
	}

	// openai v0.1.x 不提供直接的温度与最大token选项，这些参数可在调用时传入

	// 处理其他选项
	f.applyOpenAIOptions(&opts, config.Options)

	return openai.New(opts...)
}

// createDeepSeek 创建DeepSeek LLM
func (f *LLMFactory) createDeepSeek(config *LLMConfig, apiKey string) (llms.Model, error) {
	var opts []deepseek.Option

	// 设置API密钥
	if apiKey != "" {
		opts = append(opts, deepseek.WithAPIKey(apiKey))
	}

	// 设置模型
	if config.Model != "" {
		opts = append(opts, deepseek.WithModel(config.Model))
	}

	// 设置基础URL
	if config.BaseURL != "" {
		opts = append(opts, deepseek.WithBaseURL(config.BaseURL))
	}

	// DeepSeek包不支持直接设置温度和最大token数
	// 这些参数在调用时通过CallOptions传递

	// 处理其他选项
	f.applyDeepSeekOptions(&opts, config.Options)

	return deepseek.New(opts...)
}

// createKimi 创建Kimi LLM
func (f *LLMFactory) createKimi(config *LLMConfig, apiKey string) (llms.Model, error) {
	var opts []kimi.Option

	// 设置API密钥
	if apiKey != "" {
		opts = append(opts, kimi.WithToken(apiKey))
	}

	// 设置模型
	if config.Model != "" {
		opts = append(opts, kimi.WithModel(config.Model))
	}

	// 设置基础URL
	if config.BaseURL != "" {
		opts = append(opts, kimi.WithBaseURL(config.BaseURL))
	}

	// 设置温度
	if config.Temperature != nil {
		opts = append(opts, kimi.WithTemperature(*config.Temperature))
	}

	// 设置最大token数
	if config.MaxTokens != nil {
		opts = append(opts, kimi.WithMaxTokens(*config.MaxTokens))
	}

	// 处理其他选项
	f.applyKimiOptions(&opts, config.Options)

	return kimi.New(opts...)
}

// createQwen 创建Qwen LLM
func (f *LLMFactory) createQwen(config *LLMConfig, apiKey string) (llms.Model, error) {
	var opts []qwen.Option

	// 设置API密钥
	if apiKey != "" {
		opts = append(opts, qwen.WithAPIKey(apiKey))
	}

	// 设置模型
	if config.Model != "" {
		opts = append(opts, qwen.WithModel(config.Model))
	}

	// 设置基础URL
	if config.BaseURL != "" {
		opts = append(opts, qwen.WithBaseURL(config.BaseURL))
	}

	// Qwen包不支持直接设置温度和最大token数
	// 这些参数在调用时通过CallOptions传递

	// 处理其他选项
	f.applyQwenOptions(&opts, config.Options)

	return qwen.New(opts...)
}

// createAnthropic 创建Anthropic LLM
func (f *LLMFactory) createAnthropic(config *LLMConfig, apiKey string) (llms.Model, error) {
	var opts []anthropic.Option

	// 设置API密钥
	if apiKey != "" {
		opts = append(opts, anthropic.WithToken(apiKey))
	}

	// 设置模型
	if config.Model != "" {
		opts = append(opts, anthropic.WithModel(config.Model))
	}

	// 设置基础URL
	if config.BaseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(config.BaseURL))
	}

	return anthropic.New(opts...)
}

// createOllama 创建Ollama LLM
func (f *LLMFactory) createOllama(config *LLMConfig) (llms.Model, error) {
	var opts []ollama.Option

	// 设置模型
	if config.Model != "" {
		opts = append(opts, ollama.WithModel(config.Model))
	}

	// 设置基础URL
	if config.BaseURL != "" {
		opts = append(opts, ollama.WithServerURL(config.BaseURL))
	}

	return ollama.New(opts...)
}

// getDefaultAPIKey 获取默认API密钥
func (f *LLMFactory) getDefaultAPIKey(llmType string) string {
	switch llmType {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "deepseek":
		return os.Getenv("DEEPSEEK_API_KEY")
	case "kimi":
		return os.Getenv("KIMI_API_KEY")
	case "qwen":
		return os.Getenv("QWEN_API_KEY")
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	default:
		return ""
	}
}

// applyOpenAIOptions 应用OpenAI特定选项
func (f *LLMFactory) applyOpenAIOptions(opts *[]openai.Option, options map[string]interface{}) {
	if options == nil {
		return
	}

	// 处理组织ID
	if org, ok := options["organization"].(string); ok && org != "" {
		*opts = append(*opts, openai.WithOrganization(org))
	}

	// 处理API版本
	if version, ok := options["api_version"].(string); ok && version != "" {
		*opts = append(*opts, openai.WithAPIVersion(version))
	}

	// 处理嵌入模型
	if embModel, ok := options["embedding_model"].(string); ok && embModel != "" {
		*opts = append(*opts, openai.WithEmbeddingModel(embModel))
	}
}

// applyDeepSeekOptions 应用DeepSeek特定选项
func (f *LLMFactory) applyDeepSeekOptions(opts *[]deepseek.Option, options map[string]interface{}) {
	if options == nil {
		return
	}

	// 处理HTTP客户端超时
	if timeout, ok := options["timeout"].(float64); ok && timeout > 0 {
		// 可以添加超时设置选项
		_ = timeout
	}
}

// applyKimiOptions 应用Kimi特定选项
func (f *LLMFactory) applyKimiOptions(opts *[]kimi.Option, options map[string]interface{}) {
	if options == nil {
		return
	}

	// 处理HTTP客户端超时
	if timeout, ok := options["timeout"].(float64); ok && timeout > 0 {
		// 可以添加超时设置选项
		_ = timeout
	}
}

// applyQwenOptions 应用Qwen特定选项
func (f *LLMFactory) applyQwenOptions(opts *[]qwen.Option, options map[string]interface{}) {
	if options == nil {
		return
	}

	// 处理HTTP客户端超时
	if timeout, ok := options["timeout"].(float64); ok && timeout > 0 {
		// 可以添加超时设置选项
		_ = timeout
	}
}
