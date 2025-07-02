// 通义千问文本生成示例
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sjzsdu/langchaingo-cn/llms/qwen"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// 初始化通义千问客户端
	llm, err := qwen.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 使用单一提示生成内容
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"你是个什么模型？",
		llms.WithTemperature(0.7),
		llms.WithMaxTokens(1000),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("中国历史上最著名的发明:")
	fmt.Println(completion)
}
