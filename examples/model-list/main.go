package main

import (
	"fmt"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/sjzsdu/langchaingo-cn/llms/kimi"
	"github.com/sjzsdu/langchaingo-cn/llms/qwen"
	"github.com/sjzsdu/langchaingo-cn/llms/siliconflow"
	"github.com/sjzsdu/langchaingo-cn/llms/zhipu"
)

func main() {
	fmt.Println("=== LangChain-Go 中文版支持的模型列表 ===")

	// 智谱AI模型
	fmt.Println("🤖 智谱AI (ZhipuAI) 支持的模型:")
	zhipuLLM, err := zhipu.New(zhipu.WithAPIKey("dummy-key")) // 使用虚拟key仅用于演示
	if err == nil {
		models := zhipuLLM.GetModels()
		for i, model := range models {
			fmt.Printf("  %d. %s\n", i+1, model)
		}
	} else {
		fmt.Printf("  初始化失败: %v\n", err)
	}

	fmt.Println()

	// DeepSeek模型
	fmt.Println("🧠 DeepSeek 支持的模型:")
	deepseekLLM, err := deepseek.New(deepseek.WithAPIKey("dummy-key"))
	if err == nil {
		models := deepseekLLM.GetModels()
		for i, model := range models {
			fmt.Printf("  %d. %s\n", i+1, model)
		}
	} else {
		fmt.Printf("  初始化失败: %v\n", err)
	}

	fmt.Println()

	// 通义千问模型
	fmt.Println("🌟 通义千问 (Qwen) 支持的模型:")
	qwenLLM, err := qwen.New(qwen.WithAPIKey("dummy-key"))
	if err == nil {
		models := qwenLLM.GetModels()
		for i, model := range models {
			fmt.Printf("  %d. %s\n", i+1, model)
		}
	} else {
		fmt.Printf("  初始化失败: %v\n", err)
	}

	fmt.Println()

	// Kimi模型
	fmt.Println("🚀 Kimi (Moonshot) 支持的模型:")
	kimiLLM, err := kimi.New(kimi.WithToken("dummy-key"))
	if err == nil {
		models := kimiLLM.GetModels()
		for i, model := range models {
			fmt.Printf("  %d. %s\n", i+1, model)
		}
	} else {
		fmt.Printf("  初始化失败: %v\n", err)
	}

	fmt.Println()

	// 硅基流动模型
	fmt.Println("⚡ 硅基流动 (SiliconFlow) 支持的模型:")
	fmt.Println("  文本生成模型:")
	siliconflowLLM, err := siliconflow.New(siliconflow.WithAPIKey("dummy-key"))
	if err == nil {
		models := siliconflowLLM.GetModels()
		for i, model := range models {
			if i < 13 { // 前13个是文本生成模型
				fmt.Printf("    %d. %s\n", i+1, model)
			}
		}
		
		fmt.Println("  多模态模型:")
		for i, model := range models {
			if i >= 13 { // 后面的是多模态模型
				fmt.Printf("    %d. %s\n", i-12, model)
			}
		}
		
		fmt.Println("  Embedding模型:")
		embeddingModels := siliconflowLLM.GetEmbeddingModels()
		for i, model := range embeddingModels {
			fmt.Printf("    %d. %s\n", i+1, model)
		}
	} else {
		fmt.Printf("  初始化失败: %v\n", err)
	}

	fmt.Println("\n=== 使用说明 ===")
	fmt.Println("1. 每个LLM都提供了GetModels()方法来获取支持的模型列表")
	fmt.Println("2. 硅基流动还提供了GetEmbeddingModels()方法获取Embedding模型")
	fmt.Println("3. 在实际使用时，请使用真实的API密钥替换dummy-key")
	fmt.Println("4. 不同的模型有不同的性能和价格特点，请根据需求选择")

	fmt.Println("\n=== 示例代码 ===")
	fmt.Println("// 获取智谱AI支持的模型")
	fmt.Println("zhipuLLM, _ := zhipu.New(zhipu.WithAPIKey(\"your-api-key\"))")
	fmt.Println("models := zhipuLLM.GetModels()")
	fmt.Println("fmt.Println(\"支持的模型:\", models)")
}