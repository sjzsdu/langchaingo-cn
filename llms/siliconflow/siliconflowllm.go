package siliconflow

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
	TokenEnvVarName = "SILICONFLOW_API_KEY" //nolint:gosec
	ModelEnvVarName = "SILICONFLOW_MODEL"   //nolint:gosec
	// Embedding 环境变量
	EmbeddingModelEnvVarName = "SILICONFLOW_EMBEDDING_MODEL" //nolint:gosec

	// OpenAI兼容模式基础URL
	OpenAICompatibleBaseURL = "https://api.siliconflow.cn/v1"
	// 默认模型
	DefaultModel = "Qwen/Qwen2.5-72B-Instruct"
	// 默认 Embedding 模型
	DefaultEmbeddingModel = "BAAI/bge-large-zh-v1.5"
)

const (
	// ModelQwen2572B 是通义千问2.5-72B指令模型
	ModelQwen2572B = "Qwen/Qwen2.5-72B-Instruct"

	// ModelQwen257B 是通义千问2.5-7B指令模型
	ModelQwen257B = "Qwen/Qwen2.5-7B-Instruct"

	// ModelQwen2532B 是通义千问2.5-32B指令模型
	ModelQwen2532B = "Qwen/Qwen2.5-32B-Instruct"

	// ModelQwen2514B 是通义千问2.5-14B指令模型
	ModelQwen2514B = "Qwen/Qwen2.5-14B-Instruct"

	// ModelDeepSeekV25 是DeepSeek-V2.5模型
	ModelDeepSeekV25 = "deepseek-ai/DeepSeek-V2.5"

	// ModelDeepSeekR1 是DeepSeek-R1推理模型
	ModelDeepSeekR1 = "Pro/deepseek-ai/DeepSeek-R1"

	// ModelDeepSeekV3 是DeepSeek-V3模型
	ModelDeepSeekV3 = "deepseek-ai/DeepSeek-V3"

	// ModelInternLM25 是InternLM2.5-20B-Chat模型
	ModelInternLM25 = "internlm/internlm2_5-20b-chat"

	// ModelGLM49B 是GLM-4-9B-Chat模型
	ModelGLM49B = "ZHIPU/GLM-4-9B-Chat"

	// ModelYi34B 是Yi-1.5-34B-Chat模型
	ModelYi34B = "01-ai/Yi-1.5-34B-Chat"

	// ModelLlama370B 是Llama-3-70B-Instruct模型
	ModelLlama370B = "meta-llama/Meta-Llama-3-70B-Instruct"

	// ModelMistral7B 是Mistral-7B-Instruct模型
	ModelMistral7B = "mistralai/Mistral-7B-Instruct-v0.3"

	// ModelQwQ32B 是QwQ-32B-Preview推理模型
	ModelQwQ32B = "Qwen/QwQ-32B-Preview"
)

const (
	// 多模态模型
	// ModelQwenVLMax 是通义千问VL-Max多模态模型
	ModelQwenVLMax = "Qwen/Qwen2-VL-72B-Instruct"

	// ModelQwenVL7B 是通义千问VL-7B多模态模型
	ModelQwenVL7B = "Qwen/Qwen2-VL-7B-Instruct"

	// ModelInternVL2 是InternVL2-26B多模态模型
	ModelInternVL2 = "OpenGVLab/InternVL2-26B"
)

const (
	// Embedding模型
	// ModelBGELargeZh 是BGE-Large-zh向量模型
	ModelBGELargeZh = "BAAI/bge-large-zh-v1.5"

	// ModelBGEBaseZh 是BGE-Base-zh向量模型
	ModelBGEBaseZh = "BAAI/bge-base-zh-v1.5"

	// ModelBCEEmbedding 是BCE-Embedding向量模型
	ModelBCEEmbedding = "maidalun1020/bce-embedding-base_v1"

	// ModelGTEQwen2 是GTE-Qwen2-7B向量模型
	ModelGTEQwen2 = "Alibaba-NLP/gte-Qwen2-7B-instruct"
)

// LLM 是硅基流动大语言模型的实现
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

// New 创建一个新的硅基流动LLM实例
func New(opts ...Option) (*LLM, error) {
	options := defaultOptions()

	// 应用选项
	for _, opt := range opts {
		opt(&options)
	}

	// 验证API密钥
	if options.apiKey == "" {
		return nil, errors.New("API密钥不能为空，请设置SILICONFLOW_API_KEY环境变量或使用WithAPIKey选项")
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

// GetModels 返回硅基流动支持的模型列表
func (s *LLM) GetModels() []string {
	return []string{
		// 文本生成模型
		ModelQwen2572B,
		ModelQwen257B,
		ModelQwen2532B,
		ModelQwen2514B,
		ModelDeepSeekV25,
		ModelDeepSeekR1,
		ModelDeepSeekV3,
		ModelInternLM25,
		ModelGLM49B,
		ModelYi34B,
		ModelLlama370B,
		ModelMistral7B,
		ModelQwQ32B,
		// 多模态模型
		ModelQwenVLMax,
		ModelQwenVL7B,
		ModelInternVL2,
	}
}

// GetEmbeddingModels 返回硅基流动支持的Embedding模型列表
func (s *LLM) GetEmbeddingModels() []string {
	return []string{
		ModelBGELargeZh,
		ModelBGEBaseZh,
		ModelBCEEmbedding,
		ModelGTEQwen2,
	}
}

// GenerateContent 重写生成内容方法，处理推理模型的特殊返回格式
func (s *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// 硅基流动完全兼容OpenAI接口，直接调用父类方法
	return s.LLM.GenerateContent(ctx, messages, options...)
}
