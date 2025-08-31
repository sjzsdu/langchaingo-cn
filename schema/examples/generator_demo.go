package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sjzsdu/langchaingo-cn/schema"
)

func main() {
	fmt.Println("🚀 LangChainGo-CN 配置文件生成器演示")
	fmt.Println("=======================================")

	// 创建输出目录
	outputDir := "generated_configs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal("创建输出目录失败:", err)
	}

	// 创建配置生成器
	generator := schema.NewConfigGenerator(outputDir)

	// 1. 生成极简LLM配置
	fmt.Println("\n📝 生成LLM配置文件...")
	err := generator.GenerateLLMConfig(schema.LLMTemplate{
		Type:        "deepseek",
		Model:       "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   2048,
	}, "deepseek_llm.json")
	if err != nil {
		log.Printf("生成LLM配置失败: %v", err)
	}

	// 2. 生成Chain配置
	fmt.Println("\n⛓️  生成Chain配置文件...")
	err = generator.GenerateChainConfig(schema.ChainTemplate{
		Type: "conversation",
		LLMTemplate: schema.LLMTemplate{
			Type:        "kimi",
			Model:       "moonshot-v1-8k",
			Temperature: 0.7,
		},
		MemoryType:     "conversation_buffer",
		PromptTemplate: "你是一个专业的AI助手，请用中文回答用户问题：{{.input}}",
		InputVariables: []string{"input"},
	}, "kimi_chat_chain.json")
	if err != nil {
		log.Printf("生成Chain配置失败: %v", err)
	}

	// 3. 生成Agent配置
	fmt.Println("\n🤖 生成Agent配置文件...")
	err = generator.GenerateAgentConfig(schema.AgentTemplate{
		Type: "zero_shot_react",
		LLMTemplate: schema.LLMTemplate{
			Type:        "openai",
			Model:       "gpt-4",
			Temperature: 0.3,
			MaxTokens:   2048,
		},
		MemoryType: "conversation_buffer",
		MaxSteps:   5,
	}, "openai_agent.json")
	if err != nil {
		log.Printf("生成Agent配置失败: %v", err)
	}

	// 4. 生成Executor配置（新风格）
	fmt.Println("\n⚡ 生成Executor配置文件...")
	err = generator.GenerateExecutorConfig(schema.ExecutorTemplate{
		AgentTemplate: schema.AgentTemplate{
			Type: "conversational_react",
			LLMTemplate: schema.LLMTemplate{
				Type:        "qwen",
				Model:       "qwen-plus",
				Temperature: 0.5,
				MaxTokens:   1500,
			},
			MemoryType: "conversation_buffer",
			MaxSteps:   3,
		},
		MaxIterations:           8,
		ReturnIntermediateSteps: true,
	}, "qwen_executor.json")
	if err != nil {
		log.Printf("生成Executor配置失败: %v", err)
	}

	// 5. 使用快捷方法生成常用配置
	fmt.Println("\n🔥 生成常用预设配置...")

	// DeepSeek聊天配置
	err = generator.GenerateDeepSeekChatConfig("deepseek_chat.json")
	if err != nil {
		log.Printf("生成DeepSeek聊天配置失败: %v", err)
	}

	// Kimi聊天配置
	err = generator.GenerateKimiChatConfig("kimi_chat.json")
	if err != nil {
		log.Printf("生成Kimi聊天配置失败: %v", err)
	}

	// OpenAI聊天配置
	err = generator.GenerateOpenAIChatConfig("openai_chat.json")
	if err != nil {
		log.Printf("生成OpenAI聊天配置失败: %v", err)
	}

	// ReAct智能体配置
	err = generator.GenerateReactAgentConfig("deepseek", "deepseek-chat", "deepseek_react_agent.json")
	if err != nil {
		log.Printf("生成ReAct智能体配置失败: %v", err)
	}

	// 对话智能体配置
	err = generator.GenerateConversationalAgentConfig("kimi", "moonshot-v1-8k", "kimi_conversational_agent.json")
	if err != nil {
		log.Printf("生成对话智能体配置失败: %v", err)
	}

	// DeepSeek执行器配置
	err = generator.GenerateExecutorWithDeepSeek("deepseek_executor.json")
	if err != nil {
		log.Printf("生成DeepSeek执行器配置失败: %v", err)
	}

	// 6. 使用全局门面方法（最简单的方式）
	fmt.Println("\n⚡ 使用全局门面方法生成配置...")

	// 快速生成LLM配置
	err = schema.QuickGenerateLLM("anthropic", "claude-3-sonnet-20240229", "quick_anthropic.json")
	if err != nil {
		log.Printf("快速生成Anthropic配置失败: %v", err)
	}

	// 快速生成Chain配置
	err = schema.QuickGenerateChain("llm", "ollama", "llama2", "quick_ollama_chain.json")
	if err != nil {
		log.Printf("快速生成Ollama Chain配置失败: %v", err)
	}

	// 快速生成Agent配置
	err = schema.QuickGenerateAgent("zero_shot_react", "deepseek", "deepseek-chat", "quick_agent.json")
	if err != nil {
		log.Printf("快速生成Agent配置失败: %v", err)
	}

	// 快速生成Executor配置
	err = schema.QuickGenerateExecutor("conversational_react", "kimi", "moonshot-v1-8k", "quick_executor.json")
	if err != nil {
		log.Printf("快速生成Executor配置失败: %v", err)
	}

	fmt.Println("\n✨ 配置文件生成完成！")
	fmt.Printf("📁 所有配置文件已保存到目录: %s\n", outputDir)
	fmt.Println("\n💡 使用方法:")
	fmt.Println("   1. 设置相应的环境变量（如 DEEPSEEK_API_KEY、KIMI_API_KEY 等）")
	fmt.Println("   2. 使用 schema.CreateApplicationFromFile() 加载配置")
	fmt.Println("   3. 开始使用生成的组件！")

	fmt.Println("\n📋 生成的配置文件列表:")
	files := []string{
		"deepseek_llm.json",
		"kimi_chat_chain.json",
		"openai_agent.json",
		"qwen_executor.json",
		"deepseek_chat.json",
		"kimi_chat.json",
		"openai_chat.json",
		"deepseek_react_agent.json",
		"kimi_conversational_agent.json",
		"deepseek_executor.json",
		"quick_anthropic.json",
		"quick_ollama_chain.json",
		"quick_agent.json",
		"quick_executor.json",
	}

	for i, file := range files {
		fmt.Printf("   %2d. %s\n", i+1, file)
	}
}
