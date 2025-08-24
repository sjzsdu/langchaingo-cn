package schema

import (
	"fmt"
	"os"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/voyageai"
	"github.com/tmc/langchaingo/embeddings/huggingface"
	"github.com/tmc/langchaingo/embeddings/jina"
	"github.com/tmc/langchaingo/llms/openai"
)

// EmbeddingFactory Embedding组件工厂
type EmbeddingFactory struct{}

// NewEmbeddingFactory 创建Embedding工厂实例
func NewEmbeddingFactory() *EmbeddingFactory {
	return &EmbeddingFactory{}
}

// Create 根据配置创建Embedding实例
func (f *EmbeddingFactory) Create(config *EmbeddingConfig) (embeddings.Embedder, error) {
	if config == nil {
		return nil, fmt.Errorf("Embedding config is nil")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Embedding config: %w", err)
	}

	// 获取API密钥
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = f.getDefaultAPIKey(config.Type)
	}

	switch config.Type {
	case "openai":
		return f.createOpenAI(config, apiKey)
	case "voyage":
		return f.createVoyageAI(config, apiKey)
	case "huggingface":
		return f.createHuggingface(config, apiKey)
	case "jina":
		return f.createJina(config, apiKey)
	default:
		return nil, fmt.Errorf("unsupported Embedding type: %s", config.Type)
	}
}

// createOpenAI 创建OpenAI Embedding
func (f *EmbeddingFactory) createOpenAI(config *EmbeddingConfig, apiKey string) (embeddings.Embedder, error) {
	var llmOpts []openai.Option

	// 设置API密钥
	if apiKey != "" {
		llmOpts = append(llmOpts, openai.WithToken(apiKey))
	}

	// 设置模型
	if config.Model != "" {
		llmOpts = append(llmOpts, openai.WithModel(config.Model))
	}

	// 设置基础URL
	if config.BaseURL != "" {
		llmOpts = append(llmOpts, openai.WithBaseURL(config.BaseURL))
	}

	// 创建OpenAI LLM客户端
	llm, err := openai.New(llmOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI LLM: %w", err)
	}

	// 设置embedding选项
	var embeddingOpts []embeddings.Option
	if config.BatchSize != nil {
		embeddingOpts = append(embeddingOpts, embeddings.WithBatchSize(*config.BatchSize))
	}

	// 处理其他选项
	f.applyOpenAIEmbeddingOptions(&embeddingOpts, config.Options)

	// 创建嵌入器
	embedder, err := embeddings.NewEmbedder(llm, embeddingOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI embedder: %w", err)
	}

	return embedder, nil
}

// createVoyageAI 创建VoyageAI Embedding
func (f *EmbeddingFactory) createVoyageAI(config *EmbeddingConfig, apiKey string) (embeddings.Embedder, error) {
	var opts []voyageai.Option

	// 设置API密钥
	if apiKey != "" {
		opts = append(opts, voyageai.WithToken(apiKey))
	}

	// 设置模型
	if config.Model != "" {
		opts = append(opts, voyageai.WithModel(config.Model))
	}

	// 设置批处理大小
	if config.BatchSize != nil {
		opts = append(opts, voyageai.WithBatchSize(*config.BatchSize))
	}

	// 处理其他选项
	f.applyVoyageAIOptions(&opts, config.Options)

	// 创建嵌入器
	embedder, err := voyageai.NewVoyageAI(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create VoyageAI embedder: %w", err)
	}

	return embedder, nil
}

// createHuggingface 创建Huggingface Embedding
func (f *EmbeddingFactory) createHuggingface(config *EmbeddingConfig, apiKey string) (embeddings.Embedder, error) {
	var opts []huggingface.Option

	// 设置模型
	if config.Model != "" {
		opts = append(opts, huggingface.WithModel(config.Model))
	}

	// 设置批处理大小
	if config.BatchSize != nil {
		opts = append(opts, huggingface.WithBatchSize(*config.BatchSize))
	}

	// 创建嵌入器
	embedder, err := huggingface.NewHuggingface(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Huggingface embedder: %w", err)
	}

	return embedder, nil
}

// createJina 创建Jina Embedding
func (f *EmbeddingFactory) createJina(config *EmbeddingConfig, apiKey string) (embeddings.Embedder, error) {
	var opts []jina.Option

	// 设置API密钥
	if apiKey != "" {
		opts = append(opts, jina.WithAPIKey(apiKey))
	}

	// 设置模型
	if config.Model != "" {
		opts = append(opts, jina.WithModel(config.Model))
	}

	// 设置基础URL
	if config.BaseURL != "" {
		opts = append(opts, jina.WithAPIBaseURL(config.BaseURL))
	}

	// 设置批处理大小
	if config.BatchSize != nil {
		opts = append(opts, jina.WithBatchSize(*config.BatchSize))
	}

	// 创建嵌入器
	embedder, err := jina.NewJina(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jina embedder: %w", err)
	}

	return embedder, nil
}

// getDefaultAPIKey 获取默认API密钥
func (f *EmbeddingFactory) getDefaultAPIKey(embeddingType string) string {
	switch embeddingType {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "voyage":
		return os.Getenv("VOYAGEAI_API_KEY")
	case "huggingface":
		return os.Getenv("HUGGINGFACE_API_KEY")
	case "jina":
		return os.Getenv("JINA_API_KEY")
	default:
		return ""
	}
}

// applyOpenAIEmbeddingOptions 应用OpenAI Embedding特定选项
func (f *EmbeddingFactory) applyOpenAIEmbeddingOptions(opts *[]embeddings.Option, options map[string]interface{}) {
	if options == nil {
		return
	}

	// 处理是否移除换行符
	if stripNewLines, ok := options["strip_new_lines"].(bool); ok {
		*opts = append(*opts, embeddings.WithStripNewLines(stripNewLines))
	}
}

// applyVoyageAIOptions 应用VoyageAI特定选项
func (f *EmbeddingFactory) applyVoyageAIOptions(opts *[]voyageai.Option, options map[string]interface{}) {
	if options == nil {
		return
	}

	// 处理是否移除换行符
	if stripNewLines, ok := options["strip_new_lines"].(bool); ok {
		*opts = append(*opts, voyageai.WithStripNewLines(stripNewLines))
	}
}

