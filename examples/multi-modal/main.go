// 通义千问多模态示例 - 展示如何使用多模态模型处理图像和文本
//
// 本示例展示了基本多模态图片分析功能
// 运行前准备：
// 1. 确保已设置相应的API密钥环境变量
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	cnllms "github.com/sjzsdu/langchaingo-cn/llms"
	"github.com/tmc/langchaingo/llms"
)

// 定义常量
const (
	// 远程图片URL - 使用固定的图片URL
	NatureImageURL = "https://images.unsplash.com/photo-1472214103451-9374bd1c798e?w=1024&h=768" // 固定自然风景图片
)

// 运行多模态示例
func runMultiModalExample(llm llms.Model, modelName string) {
	fmt.Printf("\n===== 使用 %s 模型进行多模态分析 =====\n\n", modelName)

	// 创建上下文
	ctx := context.Background()
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
				llms.ImageURLWithDetailPart(NatureImageURL, "这张图片是什么？请简要描述一下图片中的内容。"),
			},
		},
	}

	// 添加超时控制
	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// 调用API
	response, err := llm.GenerateContent(
		timeoutCtx,
		messages,
		llms.WithMaxTokens(500),
		llms.WithTemperature(0.7),
	)

	// 错误处理
	if err != nil {
		fmt.Printf("%s, 不支持多模态：%s", modelName, err)
		fmt.Println()
		return
	}

	// 输出结果
	fmt.Println("多模态回复:")
	if len(response.Choices) > 0 {
		fmt.Println(response.Choices[0].Content)
	}
}

func main() {

	models, modelNames, err := cnllms.InitImageModels()
	if err != nil {
		log.Fatal("初始化模型失败: ", err)
	}

	// 遍历所有模型，尝试运行多模态示例
	for i, modelName := range modelNames {
		runMultiModalExample(models[i], modelName)
	}

	fmt.Println("\n=== 示例运行完成！===")
}
