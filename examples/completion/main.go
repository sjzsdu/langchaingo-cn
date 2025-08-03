package main

import (
	"context"
	"fmt"
	"log"

	cnllms "github.com/sjzsdu/langchaingo-cn/llms"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// 初始化所有模型
	models, modelNames, err := cnllms.InitTextModels()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	prompt := "你是什么模型？"

	// 依次使用每个模型生成回复
	for i, llm := range models {
		fmt.Printf("\n===== 使用 %s 模型 =====\n", modelNames[i])

		completion, err := llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			prompt,
			llms.WithTemperature(0.7),
			llms.WithMaxTokens(1000),
		)
		if err != nil {
			fmt.Printf("使用 %s 生成失败: %v\n", modelNames[i], err)
			continue
		}

		fmt.Printf("回复:\n%s\n", completion)
	}
}
