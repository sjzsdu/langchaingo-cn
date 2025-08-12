package llms

import (
	"context"
	"fmt"
	"strings"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/sjzsdu/langchaingo-cn/llms/kimi"
	"github.com/sjzsdu/langchaingo-cn/llms/qwen"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// 检查模型名称是否匹配，不区分大小写
func matchModelName(input, modelName string) bool {
	return input == "" || input == modelName || strings.EqualFold(input, modelName)
}

// 初始化所有模型并返回它们的实例
// 如果llm参数为空，则返回所有模型；否则只返回指定的模型
func InitTextModels(llm string) ([]llms.Model, []string, error) {
	// 存储所有模型实例
	var models []llms.Model
	// 存储所有模型名称
	var modelNames []string

	// 初始化DeepSeek客户端
	if matchModelName(llm, "DeepSeek") {
		deepseekLLM, err := deepseek.New()
		if err != nil {
			return nil, nil, fmt.Errorf("初始化DeepSeek失败: %w", err)
		}
		models = append(models, deepseekLLM)
		modelNames = append(modelNames, "DeepSeek")
	}

	// 初始化Qwen客户端
	if matchModelName(llm, "Qwen") {
		qwenLLM, err := qwen.New()
		if err != nil {
			return nil, nil, fmt.Errorf("初始化Qwen失败: %w", err)
		}
		models = append(models, qwenLLM)
		modelNames = append(modelNames, "Qwen")
	}

	// 初始化Kimi客户端
	if matchModelName(llm, "Kimi") {
		kimiLLM, err := kimi.New()
		if err != nil {
			return nil, nil, fmt.Errorf("初始化Kimi失败: %w", err)
		}
		models = append(models, kimiLLM)
		modelNames = append(modelNames, "Kimi")
	}

	// 如果没有找到任何模型，返回错误
	if len(models) == 0 {
		return nil, nil, fmt.Errorf("未找到指定的模型: %s", llm)
	}

	return models, modelNames, nil
}

// 初始化多模态模型并返回它们的实例
// 如果llm参数为空，则返回所有模型；否则只返回指定的模型
func InitImageModels(llm string) ([]llms.Model, []string, error) {
	// 存储所有模型实例
	var models []llms.Model
	// 存储所有模型名称
	var modelNames []string

	// 初始化DeepSeek客户端 - 使用支持多模态的模型
	if matchModelName(llm, "DeepSeek") {
		deepseekLLM, err := deepseek.New(
			deepseek.WithModel("deepseek-vision"), // 使用支持视觉的模型
		)
		if err != nil {
			return nil, nil, fmt.Errorf("初始化DeepSeek失败: %w", err)
		}
		models = append(models, deepseekLLM)
		modelNames = append(modelNames, "DeepSeek")
	}

	// 初始化Qwen客户端 - 使用支持多模态的模型
	if matchModelName(llm, "Qwen") {
		qwenLLM, err := qwen.New(
			qwen.WithModel(qwen.ModelQWenVLMax), // 使用支持视觉语言的模型
		)
		if err != nil {
			return nil, nil, fmt.Errorf("初始化Qwen失败: %w", err)
		}
		models = append(models, qwenLLM)
		modelNames = append(modelNames, "Qwen")
	}

	// 初始化Kimi客户端 - 使用支持多模态的模型
	if matchModelName(llm, "Kimi") {
		kimiLLM, err := kimi.New(
			kimi.WithModel("moonshot-vision"), // 使用支持视觉的模型
		)
		if err != nil {
			return nil, nil, fmt.Errorf("初始化Kimi失败: %w", err)
		}
		models = append(models, kimiLLM)
		modelNames = append(modelNames, "Kimi")
	}

	// 如果没有找到任何模型，返回错误
	if len(models) == 0 {
		return nil, nil, fmt.Errorf("未找到指定的模型: %s", llm)
	}

	return models, modelNames, nil
}

// hfEmbedder wraps HuggingFace LLM to provide a fixed embedding model and task.
type hfEmbedder struct {
	client *huggingface.LLM
	model  string
	task   string
}

func (h *hfEmbedder) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	return h.client.CreateEmbedding(ctx, inputTexts, h.model, h.task)
}

// InitEmbeddingModels initializes embedding-capable models.
// If llm is empty, all supported models are returned; otherwise, only the specified one.
func InitEmbeddingModels(llm string) ([]embeddings.Embedder, []string, error) {
	var models []embeddings.Embedder
	var modelNames []string

	// OpenAI: text-embedding-3-large
	if matchModelName(llm, "OpenAI") {
		openaiLLM, err := openai.New(
			openai.WithEmbeddingModel("text-embedding-3-large"),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("初始化OpenAI Embedding失败: %w", err)
		}
		e, err := embeddings.NewEmbedder(openaiLLM)
		if err != nil {
			return nil, nil, fmt.Errorf("OpenAI Embedder 创建失败: %w", err)
		}
		models = append(models, e)
		modelNames = append(modelNames, "OpenAI")
	}

	// Qwen: text-embedding-v1 (DashScope)
	if matchModelName(llm, "Qwen") {
		qwenLLM, err := qwen.New(
			qwen.WithEmbeddingModel("text-embedding-v1"),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("初始化Qwen Embedding失败: %w", err)
		}
		e, err := embeddings.NewEmbedder(qwenLLM)
		if err != nil {
			return nil, nil, fmt.Errorf("Qwen Embedder 创建失败: %w", err)
		}
		models = append(models, e)
		modelNames = append(modelNames, "Qwen")
	}

	// Ollama: bge-m3
	if matchModelName(llm, "Ollama") {
		ollamaLLM, err := ollama.New(
			ollama.WithModel("bge-m3"),
			ollama.WithRunnerEmbeddingOnly(true),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("初始化Ollama Embedding失败: %w", err)
		}
		e, err := embeddings.NewEmbedder(ollamaLLM)
		if err != nil {
			return nil, nil, fmt.Errorf("Ollama Embedder 创建失败: %w", err)
		}
		models = append(models, e)
		modelNames = append(modelNames, "Ollama")
	}

	// HuggingFace: sentence-transformers/all-MiniLM-L6-v2 via feature-extraction
	if matchModelName(llm, "HuggingFace") {
		hfLLM, err := huggingface.New()
		if err != nil {
			return nil, nil, fmt.Errorf("初始化HuggingFace Embedding失败: %w", err)
		}
		hf := &hfEmbedder{client: hfLLM, model: "sentence-transformers/all-MiniLM-L6-v2", task: "feature-extraction"}
		e, err := embeddings.NewEmbedder(hf)
		if err != nil {
			return nil, nil, fmt.Errorf("HuggingFace Embedder 创建失败: %w", err)
		}
		models = append(models, e)
		modelNames = append(modelNames, "HuggingFace")
	}

	if len(models) == 0 {
		return nil, nil, fmt.Errorf("未找到指定的Embedding模型: %s", llm)
	}

	return models, modelNames, nil
}
