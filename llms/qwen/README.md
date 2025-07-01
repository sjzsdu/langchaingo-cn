# 通义千问 LLM

本包提供了阿里云通义千问大语言模型的Go语言客户端实现，支持通过API调用通义千问模型进行文本生成、多模态内容处理和工具调用等功能。

## 功能特性

- 支持文本生成和聊天补全
- 支持多模态内容处理（图像输入）
- 支持工具调用
- 支持流式输出
- 支持OpenAI兼容模式和DashScope原生API

## 使用方法

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms/qwen"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	// 创建通义千问LLM客户端
	llm, err := qwen.New(
		qwen.WithAPIKey(os.Getenv("QWEN_API_KEY")),
		qwen.WithModel("qwen-plus"), // 可选：指定模型，默认为qwen-turbo
	)
	if err != nil {
		fmt.Printf("创建通义千问客户端失败: %v\n", err)
		return
	}

	// 调用模型生成文本
	ctx := context.Background()
	response, err := llm.Call(ctx, "你好，请介绍一下自己")
	if err != nil {
		fmt.Printf("调用通义千问失败: %v\n", err)
		return
	}

	fmt.Printf("通义千问回复: %s\n", response)

	// 多轮对话示例
	messages := []schema.ChatMessage{
		schema.HumanChatMessage{Content: "你好，请介绍一下自己"},
		schema.AIChatMessage{Content: "我是通义千问，阿里云推出的大语言模型。有什么可以帮助你的？"},
		schema.HumanChatMessage{Content: "你能做什么？"},
	}

	response, err = llm.GenerateContent(ctx, messages, nil)
	if err != nil {
		fmt.Printf("多轮对话失败: %v\n", err)
		return
	}

	fmt.Printf("多轮对话回复: %s\n", response.Content)
}
```

## 环境变量配置

可以通过环境变量配置API密钥和其他参数：

```bash
export QWEN_API_KEY="your-api-key"
```

## 参考文档

- [通义千问API文档](https://help.aliyun.com/zh/model-studio/developer-reference/use-qwen-by-calling-api)
- [DashScope SDK文档](https://help.aliyun.com/zh/dashscope/developer-reference/install-dashscope-sdk)