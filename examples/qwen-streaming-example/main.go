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

	// 创建聊天消息
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "你是一个专业的中国历史学家，擅长用通俗易懂的语言讲解历史事件和人物。"),
		llms.TextParts(llms.ChatMessageTypeHuman, "请介绍一下中国四大发明及其历史意义。"),
	}

	// 使用流式输出生成内容
	fmt.Println("开始流式生成内容...\n")

	// 使用StreamingGenerateContent方法获取流式响应
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
