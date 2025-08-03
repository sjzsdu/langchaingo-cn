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

	// 创建聊天消息
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "你是什么公司开发的？"),
	}

	// 依次使用每个模型进行流式生成
	for i, llm := range models {
		fmt.Printf("\n===== 使用 %s 模型流式生成 =====\n\n", modelNames[i])

		// 使用StreamingFunc接收流式输出
		_, err = llm.GenerateContent(
			ctx,
			content,
			llms.WithMaxTokens(1000),
			llms.WithTemperature(0.7),
			llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
				fmt.Print(string(chunk))
				return nil
			}),
		)
		if err != nil {
			fmt.Printf("\n使用 %s 流式生成失败: %v\n", modelNames[i], err)
			continue
		}

		fmt.Println("\n\n流式生成完成")
	}
}
