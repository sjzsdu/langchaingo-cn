package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sjzsdu/langchaingo-cn/schema"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	// 示例1：从JSON配置文件创建应用
	fmt.Println("=== 示例1：从JSON配置文件创建应用 ===")
	app, err := schema.CreateApplicationFromFile("schema/examples/simple_chat.json")
	if err != nil {
		log.Printf("创建应用失败: %v", err)
	} else {
		fmt.Printf("成功创建应用，包含 %d 个LLM，%d 个Memory，%d 个Chain\n",
			len(app.LLMs), len(app.Memories), len(app.Chains))
	}

	// 示例2：从JSON字符串创建应用
	fmt.Println("\n=== 示例2：从JSON字符串创建应用 ===")
	jsonConfig := `{
		"llms": {
			"chat_llm": {
				"type": "deepseek",
				"model": "deepseek-chat",
				"api_key": "${DEEPSEEK_API_KEY}",
				"temperature": 0.7
			}
		},
		"memories": {
			"chat_memory": {
				"type": "conversation_buffer",
				"max_messages": 5
			}
		},
		"chains": {
			"chat_chain": {
				"type": "conversation",
				"llm_ref": "chat_llm",
				"memory_ref": "chat_memory"
			}
		}
	}`

	app2, err := schema.CreateApplicationFromJSON(jsonConfig)
	if err != nil {
		log.Printf("创建应用失败: %v", err)
	} else {
		fmt.Printf("成功创建应用，包含 %d 个LLM，%d 个Memory，%d 个Chain\n",
			len(app2.LLMs), len(app2.Memories), len(app2.Chains))

		// 使用创建的链进行对话
		if chatChain, exists := app2.Chains["chat_chain"]; exists {
			ctx := context.Background()
			result, err := chains.Run(ctx, chatChain, "你好，请介绍一下自己")
			if err != nil {
				log.Printf("对话失败: %v", err)
			} else {
				fmt.Printf("AI回复: %s\n", result)
			}
		}
	}

	// 示例3：创建单个组件
	fmt.Println("\n=== 示例3：创建单个组件 ===")

	// 创建LLM
	llmConfig := &schema.LLMConfig{
		Type:        "openai",
		Model:       "gpt-3.5-turbo",
		APIKey:      "${OPENAI_API_KEY}",
		Temperature: floatPtr(0.8),
		MaxTokens:   intPtr(1500),
	}

	llm, err := schema.CreateLLMFromConfig(llmConfig)
	if err != nil {
		log.Printf("创建LLM失败: %v", err)
	} else {
		fmt.Println("成功创建OpenAI LLM")

		// 测试LLM（仅当设置了有效 OPENAI_API_KEY）
		if k := os.Getenv("OPENAI_API_KEY"); k != "" && k != "${OPENAI_API_KEY}" {
			ctx := context.Background()
			response, err := llm.Call(ctx, "Hello, how are you?")
			if err != nil {
				log.Printf("LLM调用失败: %v", err)
			} else {
				fmt.Printf("LLM回复: %s\n", response)
			}
		} else {
			fmt.Println("跳过OpenAI调用演示：未设置有效 OPENAI_API_KEY")
		}
	}

	// 创建Memory
	memoryConfig := &schema.MemoryConfig{
		Type:           "conversation_buffer",
		MaxMessages:    intPtr(10),
		ReturnMessages: boolPtr(true),
	}

	_, err = schema.CreateMemoryFromConfig(memoryConfig, nil)
	if err != nil {
		log.Printf("创建Memory失败: %v", err)
	} else {
		fmt.Println("成功创建会话缓冲Memory")
	}

	// 创建Prompt
	promptConfig := &schema.PromptConfig{
		Type:           "prompt_template",
		Template:       "请以专业的{{.role}}身份回答：{{.question}}",
		InputVariables: []string{"role", "question"},
	}

	prompt, err := schema.CreatePromptFromConfig(promptConfig)
	if err != nil {
		log.Printf("创建Prompt失败: %v", err)
	} else {
		fmt.Println("成功创建Prompt模板")

		// 格式化Prompt (仅当为普通 PromptTemplate 时)
		if pt, ok := prompt.(prompts.PromptTemplate); ok {
			formatted, err := pt.Format(map[string]any{
				"role":     "软件工程师",
				"question": "如何优化数据库查询性能？",
			})
			if err != nil {
				log.Printf("Prompt格式化失败: %v", err)
			} else {
				fmt.Printf("格式化后的Prompt: %s\n", formatted)
			}
		} else {
			fmt.Println("当前Prompt不是普通模板，跳过直接格式化演示")
		}
	}

	// 示例4：配置验证
	fmt.Println("\n=== 示例4：配置验证 ===")

	// 加载配置并验证
	config, err := schema.LoadConfigFromFile("schema/examples/complex_app.json")
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	} else {
		// 执行详细验证
		result := schema.ValidateConfig(config)
		fmt.Printf("配置验证结果:\n%s\n", result.String())

		if result.HasErrors() {
			fmt.Println("发现配置错误，请修复后重试")
		} else {
			fmt.Println("配置验证通过！")
		}
	}

	// 示例5：错误处理
	fmt.Println("\n=== 示例5：错误处理 ===")

	invalidConfig := `{
		"llms": {
			"invalid_llm": {
				"type": "unknown_type",
				"model": "test"
			}
		}
	}`

	_, err = schema.CreateApplicationFromJSON(invalidConfig)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)

		// 检查错误类型
		if schemaErr, ok := err.(*schema.SchemaError); ok {
			fmt.Printf("错误类型: %s\n", schemaErr.Type)
			fmt.Printf("错误路径: %s\n", schemaErr.Path)
		}
	}
}

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
