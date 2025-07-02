package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// 初始化DeepSeek客户端
	llm, err := deepseek.New(
		deepseek.WithModel("deepseek-vl2-small"), // 使用DeepSeek的视觉语言模型
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 获取图像文件的绝对路径
	imageFilePath, err := filepath.Abs("./image.svg")
	if err != nil {
		log.Fatal(err)
	}

	// 检查图像文件是否存在
	if _, err := os.Stat(imageFilePath); os.IsNotExist(err) {
		log.Fatalf("图像文件不存在: %s\n请确保在示例目录中放置了名为'image.svg'的图像文件", imageFilePath)
	}

	// 创建图像URL内容
	imageURL := llms.ImageURLContent{
		URL:    fmt.Sprintf("file://%s", imageFilePath),
		Detail: "high", // 可选值: "low", "high"
	}

	// 创建多模态消息内容
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "你是一个专业的图像分析助手，擅长分析图像内容并提供详细描述。",
				},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "请详细描述这张图片中的内容，包括主要对象、场景、颜色和可能的活动。",
				},
				imageURL,
			},
		},
	}

	// 生成内容
	completion, err := llm.GenerateContent(
		ctx,
		messages,
		llms.WithMaxTokens(1000),
		llms.WithTemperature(0.7),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 输出结果
	fmt.Println("图像分析结果:")
	if len(completion.Choices) > 0 {
		fmt.Println(completion.Choices[0].Content)
	}
}
