package llms

import (
	"fmt"
	"strings"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/sjzsdu/langchaingo-cn/llms/kimi"
	"github.com/sjzsdu/langchaingo-cn/llms/qwen"
	"github.com/tmc/langchaingo/llms"
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
