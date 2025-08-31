package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sjzsdu/langchaingo-cn/schema"
)

func main() {
	fmt.Println("=== ExecutorUsageConfig 创建 Executor 功能演示 ===")

	// 示例 1: 从 JSON 字符串直接创建
	fmt.Println("\n1. 从 JSON 字符串创建 Executor:")
	jsonConfig := `{
		"agent": {
			"type": "zero_shot_react",
			"chain": {
				"type": "llm",
				"llm": {
					"type": "deepseek",
					"model": "deepseek-chat",
					"api_key": "${DEEPSEEK_API_KEY}"
				}
			},
			"output_key": "result"
		},
		"max_iterations": 3,
		"return_intermediate_steps": true
	}`

	executor1, err := schema.CreateExecutorFromJSON(jsonConfig)
	if err != nil {
		log.Printf("创建失败: %v", err)
	} else {
		fmt.Printf("✓ 成功创建 Executor: %T\n", executor1)
	}

	// 示例 2: 使用 ExecutorUsageConfig 方法
	fmt.Println("\n2. 使用 ExecutorUsageConfig.CreateExecutor() 方法:")
	config, err := schema.LoadExecutorUsageConfigFromJSON(jsonConfig)
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		return
	}

	executor2, err := config.CreateExecutor()
	if err != nil {
		log.Printf("创建失败: %v", err)
	} else {
		fmt.Printf("✓ 成功创建 Executor: %T\n", executor2)
	}

	// 示例 3: 从文件创建（如果文件存在）
	fmt.Println("\n3. 从文件创建 Executor:")
	filename := "/Users/juzhongsun/Codes/gos/langchaingo-cn/schema/examples/simple_executor.json"
	if _, err := os.Stat(filename); err == nil {
		executor3, err := schema.CreateExecutorFromFile(filename)
		if err != nil {
			log.Printf("从文件创建失败: %v", err)
		} else {
			fmt.Printf("✓ 从文件成功创建 Executor: %T\n", executor3)
		}
	} else {
		fmt.Printf("⚠ 文件不存在: %s\n", filename)
	}

	// 示例 4: 使用全局工厂方法
	fmt.Println("\n4. 使用全局工厂方法:")
	executor4, err := schema.CreateExecutorFromUsageJSON(jsonConfig)
	if err != nil {
		log.Printf("创建失败: %v", err)
	} else {
		fmt.Printf("✓ 成功创建 Executor: %T\n", executor4)
	}

	// 演示基本使用（不会真正调用，因为没有有效的 API key）
	fmt.Println("\n5. 演示 Executor 基本接口:")
	if executor1 != nil {
		fmt.Println("Executor 已创建，可以调用:")
		fmt.Println("  - executor.Call(ctx, inputs)")
		fmt.Println("  - executor.Run(ctx, inputs)")
		
		// 示例调用（会因为无效 API key 而失败，但展示了接口）
		inputs := map[string]any{
			"input": "你好，世界！",
		}
		
		fmt.Printf("尝试调用: executor.Call(ctx, %v)\n", inputs)
		_, err := executor1.Call(context.Background(), inputs)
		if err != nil {
			fmt.Printf("预期错误（缺少有效 API key）: %v\n", err)
		}
	}

	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("\n现在 ExecutorUsageConfig 支持以下创建方法:")
	fmt.Println("  - config.CreateExecutor()")
	fmt.Println("  - schema.CreateExecutorFromJSON(jsonStr)")
	fmt.Println("  - schema.CreateExecutorFromFile(filename)")
	fmt.Println("  - schema.CreateExecutorFromUsageJSON(jsonStr)")
	fmt.Println("  - schema.CreateExecutorFromUsageFile(filename)")
}
