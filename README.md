# LangChainGo-CN

[![Go Reference](https://pkg.go.dev/badge/github.com/sjzsdu/langchaingo-cn.svg)](https://pkg.go.dev/github.com/sjzsdu/langchaingo-cn)

LangChainGo-CN 是一个基于 [LangChainGo](https://github.com/tmc/langchaingo) 的扩展库，专为中文开发者提供对国内主流大语言模型的支持。该库提供了简单统一的接口，让您可以轻松地与多种中文大语言模型进行交互。

## 支持的模型

目前支持以下模型：

- **DeepSeek**：深度求索AI的大语言模型
- **Qwen**：阿里云通义千问大语言模型
- **Kimi**：Moonshot AI的Kimi大语言模型

同时，通过底层的LangChainGo库，也支持：

- **OpenAI**：包括GPT系列模型
- **Anthropic**：Claude系列模型
- **Ollama**：本地部署的开源模型

## 功能特性

- **统一接口**：提供一致的API接口，轻松切换不同的模型
- **多模态支持**：支持图像和文本的多模态输入和处理
- **流式响应**：支持流式生成，实时获取模型响应
- **工具调用**：支持函数调用功能，让模型能够调用外部工具和API

## 安装

```bash
go get github.com/sjzsdu/langchaingo-cn
```

## 快速开始

### 基本文本生成

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cnllms "github.com/sjzsdu/langchaingo-cn/llms"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// 初始化所有模型或指定模型
	models, modelNames, err := cnllms.InitTextModels("") // 传入空字符串初始化所有模型
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
```

### 工具调用示例

```go
package main

import (
	"context"
	"fmt"
	"log"

	cnllms "github.com/sjzsdu/langchaingo-cn/llms"
	"github.com/tmc/langchaingo/llms"
)

// 定义工具接口
type Tool interface {
	// 获取工具定义
	GetDefinition() llms.Tool
	// 执行工具
	Execute(args map[string]interface{}) (interface{}, error)
}

// 天气工具实现
type WeatherTool struct{}

// 获取工具定义
func (w *WeatherTool) GetDefinition() llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "获取指定位置和日期的天气信息",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "城市名称，如北京、上海等",
					},
					"date": map[string]interface{}{
						"type":        "string",
						"description": "查询日期，格式为YYYY-MM-DD",
					},
				},
				"required": []string{"location"},
			},
		},
	}
}

// 主函数
func main() {
	// 初始化模型
	models, modelNames, err := cnllms.InitTextModels("")
	if err != nil {
		log.Fatal(err)
	}

	// 创建工具
	weatherTool := &WeatherTool{}

	// 使用模型和工具
	for i, model := range models {
		// 处理工具调用
		handleToolCalls(model, modelNames[i], weatherTool)
	}
}
```

### 多模态示例

```go
package main

import (
	"context"
	"fmt"
	"log"

	cnllms "github.com/sjzsdu/langchaingo-cn/llms"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// 初始化多模态模型
	models, modelNames, err := cnllms.InitImageModels("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	
	// 创建多模态消息内容
	imageURL := "https://example.com/image.jpg"
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
				llms.ImageURLWithDetailPart(imageURL, "这张图片是什么？请简要描述一下图片中的内容。"),
			},
		},
	}

	// 使用模型分析图像
	for i, model := range models {
		fmt.Printf("\n===== 使用 %s 模型进行多模态分析 =====\n\n", modelNames[i])
		
		response, err := llms.GenerateContent(ctx, model, messages)
		if err != nil {
			fmt.Printf("使用 %s 分析失败: %v\n", modelNames[i], err)
			continue
		}
		
		fmt.Printf("分析结果:\n%s\n", response.Content)
	}
}
```

## 环境变量配置

使用前需要设置相应的API密钥环境变量：

- DeepSeek: `DEEPSEEK_API_KEY`
- Qwen: `QWEN_API_KEY`
- Kimi: `KIMI_API_KEY`
- OpenAI: `OPENAI_API_KEY`
- Anthropic: `ANTHROPIC_API_KEY`

## 高级配置

可以通过以下参数自定义模型行为：

- `WithTemperature`: 控制生成文本的随机性
- `WithMaxTokens`: 设置生成文本的最大长度
- `WithTopP`: 控制生成文本的多样性
- `WithTopK`: 控制生成文本的多样性（仅部分模型支持）

## 贡献

欢迎提交问题和拉取请求！

## 许可证

本项目采用 [MIT 许可证](LICENSE)。