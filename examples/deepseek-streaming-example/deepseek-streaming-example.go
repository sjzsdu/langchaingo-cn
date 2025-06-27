package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// 初始化DeepSeek客户端
	llm, err := deepseek.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 创建聊天消息
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "你是一个专业的科普助手，擅长用通俗易懂的语言解释复杂概念。"),
		llms.TextParts(llms.ChatMessageTypeHuman, "请解释量子纠缠现象以及它对量子计算的重要性。"),
	}

	// 使用流式输出生成内容
	fmt.Println("开始流式生成内容...\n")

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
		log.Fatal(err)
	}

	fmt.Println("\n\n流式生成完成")
}
