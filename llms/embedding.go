package llms

import (
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/huggingface"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"

	// add qwen
	"github.com/sjzsdu/langchaingo-cn/llms/qwen"
)

// EmbeddingType 表示向量模型类型
type EmbeddingType string

// 支持的Embedding类型常量
const (
	OpenAIEmbedding      EmbeddingType = "openai"
	OllamaEmbedding      EmbeddingType = "ollama"
	HuggingFaceEmbedding EmbeddingType = "huggingface"
	// qwen
	QwenEmbedding EmbeddingType = "qwen"
)

// ErrUnsupportedEmbeddingType 表示不支持的Embedding类型错误
var ErrUnsupportedEmbeddingType = errors.New("不支持的Embedding类型")

// CreateEmbedding 创建指定类型的Embedding实例
// params 参考：
// - 通用："model"、"base_url"、"api_key" 等
// - OpenAI："organization"、"api_type"("openai"|"azure"|"azure_ad")、"api_version"、"embedding_model"
// - Ollama："server_url"(默认 http://localhost:11434)、"model"(默认 bge-m3)
// - HuggingFace："api_key"、"model"(默认 sentence-transformers/all-MiniLM-L6-v2)、"task"(默认 feature-extraction)
// - Qwen："api_key"、"model"(默认 qwen-max)、"embedding_model"(默认 text-embedding-v1)
func CreateEmbedding(embType EmbeddingType, params map[string]interface{}) (embeddings.Embedder, error) {
	switch embType {
	case OpenAIEmbedding:
		return createOpenAIEmbedding(params)
	case OllamaEmbedding:
		return createOllamaEmbedding(params)
	case HuggingFaceEmbedding:
		return createHuggingFaceEmbedding(params)
	case QwenEmbedding:
		return createQwenEmbedding(params)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedEmbeddingType, embType)
	}
}

func createOpenAIEmbedding(params map[string]interface{}) (embeddings.Embedder, error) {
	opts := []openai.Option{}

	if apiKey, ok := params["api_key"].(string); ok && apiKey != "" {
		opts = append(opts, openai.WithToken(apiKey))
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

	if embModel, ok := params["embedding_model"].(string); ok && embModel != "" {
		opts = append(opts, openai.WithEmbeddingModel(embModel))
	} else {
		// 默认使用最新常用的 embedding 模型
		opts = append(opts, openai.WithEmbeddingModel("text-embedding-3-large"))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}
	return embeddings.NewEmbedder(llm)
}

func createOllamaEmbedding(params map[string]interface{}) (embeddings.Embedder, error) {
	opts := []ollama.Option{}

	serverURL, _ := params["server_url"].(string)
	if serverURL == "" {
		serverURL = "http://localhost:11434"
	}
	opts = append(opts, ollama.WithServerURL(serverURL))

	model, _ := params["model"].(string)
	if model == "" {
		model = "bge-m3"
	}
	opts = append(opts, ollama.WithModel(model))
	opts = append(opts, ollama.WithRunnerEmbeddingOnly(true))

	llm, err := ollama.New(opts...)
	if err != nil {
		return nil, err
	}
	return embeddings.NewEmbedder(llm)
}

func createHuggingFaceEmbedding(params map[string]interface{}) (embeddings.Embedder, error) {
	opts := []huggingface.Option{}
	if token, ok := params["api_key"].(string); ok && token != "" {
		opts = append(opts, huggingface.WithToken(token))
	}

	hf, err := huggingface.New(opts...)
	if err != nil {
		return nil, err
	}

	model := "sentence-transformers/all-MiniLM-L6-v2"
	if m, ok := params["model"].(string); ok && m != "" {
		model = m
	}
	task := "feature-extraction"
	if t, ok := params["task"].(string); ok && t != "" {
		task = t
	}

	client := &hfEmbedder{client: hf, model: model, task: task}
	return embeddings.NewEmbedder(client)
}

func createQwenEmbedding(params map[string]interface{}) (embeddings.Embedder, error) {
	opts := []qwen.Option{}

	if apiKey, ok := params["api_key"].(string); ok && apiKey != "" {
		opts = append(opts, qwen.WithAPIKey(apiKey))
	}
	if baseURL, ok := params["base_url"].(string); ok && baseURL != "" {
		opts = append(opts, qwen.WithBaseURL(baseURL))
	}
	if model, ok := params["model"].(string); ok && model != "" {
		opts = append(opts, qwen.WithModel(model))
	}
	if embModel, ok := params["embedding_model"].(string); ok && embModel != "" {
		opts = append(opts, qwen.WithEmbeddingModel(embModel))
	}

	llm, err := qwen.New(opts...)
	if err != nil {
		return nil, err
	}
	return embeddings.NewEmbedder(llm)
}
