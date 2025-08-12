package llms

import (
	"errors"
	"fmt"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/sjzsdu/langchaingo-cn/llms/kimi"
	"github.com/sjzsdu/langchaingo-cn/llms/qwen"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// LLMType 表示LLM的类型
type LLMType string

// 支持的LLM类型常量
const (
	DeepSeekLLM  LLMType = "deepseek"
	KimiLLM      LLMType = "kimi"
	QwenLLM      LLMType = "qwen"
	AnthropicLLM LLMType = "anthropic"
	OpenAILLM    LLMType = "openai"
	OllamaLLM    LLMType = "ollama"
)

// ErrUnsupportedLLMType 表示不支持的LLM类型错误
var ErrUnsupportedLLMType = errors.New("不支持的LLM类型")

// ErrMissingRequiredParam 表示缺少必要参数错误
var ErrMissingRequiredParam = errors.New("缺少必要参数")

// CreateLLM 创建指定类型的LLM实例
// llmType: LLM类型
// params: 创建LLM所需的参数，不同类型的LLM需要不同的参数
//
// 常用参数：
// - "api_key": API密钥（大多数LLM都需要）
// - "server_url": 服务器URL（Ollama使用，默认为"http://localhost:11434"）
//
// 可选参数：
// - "model": 模型名称
// - "base_url": API基础URL
// - "temperature": 温度参数
// - "top_p": Top-P参数
// - "top_k": Top-K参数（仅Qwen支持）
// - "max_tokens": 最大生成令牌数
// - "use_openai_compatible": 是否使用OpenAI兼容模式（仅Qwen支持）
// - "organization": 组织ID（仅OpenAI支持）
// - "api_type": API类型（仅OpenAI支持，可选值："openai"、"azure"、"azure_ad"）
// - "api_version": API版本（仅OpenAI支持，默认为"2023-05-15"）
// - "format": 输出格式（仅Ollama支持，可选值："json"）
// - "system": 系统提示（仅Ollama支持）
func CreateLLM(llmType LLMType, params map[string]interface{}) (llms.Model, error) {
	switch llmType {
	case DeepSeekLLM:
		return createDeepSeekLLM(params)
	case KimiLLM:
		return createKimiLLM(params)
	case QwenLLM:
		return createQwenLLM(params)
	case AnthropicLLM:
		return createAnthropicLLM(params)
	case OpenAILLM:
		return createOpenAILLM(params)
	case OllamaLLM:
		return createOllamaLLM(params)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLLMType, llmType)
	}
}

// createDeepSeekLLM 创建DeepSeek LLM实例
func createDeepSeekLLM(params map[string]interface{}) (*deepseek.LLM, error) {
	// 必填校验
	apiKey, _ := params["api_key"].(string)
	if apiKey == "" {
		return nil, ErrMissingRequiredParam
	}

	// 构建选项
	opts := []deepseek.Option{deepseek.WithAPIKey(apiKey)}

	if model, ok := params["model"].(string); ok && model != "" {
		opts = append(opts, deepseek.WithModel(model))
	}

	if baseURL, ok := params["base_url"].(string); ok && baseURL != "" {
		opts = append(opts, deepseek.WithBaseURL(baseURL))
	}

	// 创建LLM实例
	return deepseek.New(opts...)
}

// createAnthropicLLM 创建Anthropic LLM实例
func createAnthropicLLM(params map[string]interface{}) (llms.Model, error) {
	// 构建选项
	opts := []anthropic.Option{}

	// 添加参数
	if apiKey, ok := params["api_key"].(string); ok && apiKey != "" {
		opts = append(opts, anthropic.WithToken(apiKey))
	}

	if model, ok := params["model"].(string); ok && model != "" {
		opts = append(opts, anthropic.WithModel(model))
	}

	if baseURL, ok := params["base_url"].(string); ok && baseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(baseURL))
	}

	// 创建LLM实例
	return anthropic.New(opts...)
}

// createKimiLLM 创建Kimi LLM实例
func createKimiLLM(params map[string]interface{}) (*kimi.LLM, error) {
	// 必填校验
	apiKey, _ := params["api_key"].(string)
	if apiKey == "" {
		return nil, ErrMissingRequiredParam
	}

	// 构建选项
	opts := []kimi.Option{kimi.WithToken(apiKey)}

	if model, ok := params["model"].(string); ok && model != "" {
		opts = append(opts, kimi.WithModel(model))
	}

	if baseURL, ok := params["base_url"].(string); ok && baseURL != "" {
		opts = append(opts, kimi.WithBaseURL(baseURL))
	}

	if temperature, ok := params["temperature"].(float64); ok {
		opts = append(opts, kimi.WithTemperature(temperature))
	}

	if topP, ok := params["top_p"].(float64); ok {
		opts = append(opts, kimi.WithTopP(topP))
	}

	if maxTokens, ok := params["max_tokens"].(int); ok {
		opts = append(opts, kimi.WithMaxTokens(maxTokens))
	}

	// 创建LLM实例
	return kimi.New(opts...)
}

// createQwenLLM 创建Qwen LLM实例
func createQwenLLM(params map[string]interface{}) (*qwen.LLM, error) {
	// 必填校验
	apiKey, _ := params["api_key"].(string)
	if apiKey == "" {
		return nil, ErrMissingRequiredParam
	}

	// 构建选项
	opts := []qwen.Option{qwen.WithAPIKey(apiKey)}

	if model, ok := params["model"].(string); ok && model != "" {
		opts = append(opts, qwen.WithModel(model))
	}

	if baseURL, ok := params["base_url"].(string); ok && baseURL != "" {
		opts = append(opts, qwen.WithBaseURL(baseURL))
	}
	// 创建LLM实例
	return qwen.New(opts...)
}

// createOpenAILLM 创建OpenAI LLM实例
func createOpenAILLM(params map[string]interface{}) (llms.Model, error) {
	// 构建选项
	opts := []openai.Option{}

	// 添加参数
	if apiKey, ok := params["api_key"].(string); ok && apiKey != "" {
		opts = append(opts, openai.WithToken(apiKey))
	}

	if model, ok := params["model"].(string); ok && model != "" {
		opts = append(opts, openai.WithModel(model))
	}

	if baseURL, ok := params["base_url"].(string); ok && baseURL != "" {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}

	if organization, ok := params["organization"].(string); ok && organization != "" {
		opts = append(opts, openai.WithOrganization(organization))
	}

	if apiType, ok := params["api_type"].(string); ok && apiType != "" {
		var apiTypeEnum openai.APIType
		switch apiType {
		case "azure":
			apiTypeEnum = openai.APITypeAzure
		case "azure_ad":
			apiTypeEnum = openai.APITypeAzureAD
		default:
			apiTypeEnum = openai.APITypeOpenAI
		}
		opts = append(opts, openai.WithAPIType(apiTypeEnum))
	}

	if apiVersion, ok := params["api_version"].(string); ok && apiVersion != "" {
		opts = append(opts, openai.WithAPIVersion(apiVersion))
	}

	// 创建LLM实例
	return openai.New(opts...)
}

// createOllamaLLM 创建Ollama LLM实例
func createOllamaLLM(params map[string]interface{}) (llms.Model, error) {
	// 构建选项
	opts := []ollama.Option{}

	// 设置服务器URL（默认为"http://localhost:11434"）
	serverURL, ok := params["server_url"].(string)
	if !ok || serverURL == "" {
		serverURL = "http://localhost:11434"
	}
	opts = append(opts, ollama.WithServerURL(serverURL))

	// 添加可选参数
	if model, ok := params["model"].(string); ok && model != "" {
		opts = append(opts, ollama.WithModel(model))
	}

	// 注意：Ollama 的温度设置是通过 Options 结构体中的 Temperature 字段设置的
	// 目前 ollama 包没有提供直接设置温度的选项函数

	if format, ok := params["format"].(string); ok && format == "json" {
		opts = append(opts, ollama.WithFormat("json"))
	}

	if system, ok := params["system"].(string); ok && system != "" {
		opts = append(opts, ollama.WithSystemPrompt(system))
	}

	// 创建LLM实例
	return ollama.New(opts...)
}
