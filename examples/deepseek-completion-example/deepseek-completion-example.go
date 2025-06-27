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

	// 使用单一提示生成内容
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"中国历史上最著名的发明是什么？请列举五个并简要说明其重要性。",
		llms.WithTemperature(0.7),
		llms.WithMaxTokens(1000),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("中国历史上最著名的发明:")
	fmt.Println(completion)
}
