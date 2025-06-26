# LangChainGo-CN

这个项目是对[LangChainGo](https://github.com/tmc/langchaingo)的扩展，主要实现了中国大模型厂商的接口集成。

## 支持的模型

- DeepSeek
- 更多模型正在添加中...

## 安装

```bash
go get github.com/sjzsdu/langchaingo-cn
```

## 使用方法

### DeepSeek

#### 基本调用

```go
package main

import (
	"context"
	"fmt"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
)

func main() {
	llm, err := deepseek.New(
		deepseek.WithAPIKey("your-api-key"),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	response, err := llm.Call(ctx, "你好，请介绍一下你自己")
	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
```

#### 使用GenerateContent方法

```go
package main

import (
	"context"
	"fmt"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	model, err := deepseek.New(
		deepseek.WithAPIKey("your-api-key"),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 创建消息内容
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "你是一个专业的Go语言助手"},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Go语言中，如何实现接口？"},
			},
		},
	}

	// 设置选项
	options := []llms.CallOption{
		llms.WithTemperature(0.7),
		llms.WithMaxTokens(1000),
	}

	// 调用GenerateContent方法
	response, err := model.GenerateContent(ctx, messages, options...)
	if err != nil {
		panic(err)
	}

	fmt.Println(response.Choices[0].Content)
}
```

## 许可证

MIT