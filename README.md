# LangChainGo-CN

这个项目是对[LangChainGo](https://github.com/tmc/langchaingo)的扩展，主要实现了中国大模型厂商的接口集成。

## 支持的模型

- DeepSeek
- QWen
- 更多模型正在添加中...

## 支持的功能

### DeepSeek

- 基本文本生成
- 多轮对话
- 流式输出（Streaming）
- 工具调用（Function Calling）
- 多模态输入（图像输入）

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

#### 使用工具调用（Function Calling）

DeepSeek模型支持工具调用（Function Calling）功能，可以让模型根据用户的输入调用预定义的工具函数，并根据工具函数的返回结果生成最终回复。以下是一个使用天气查询工具的完整示例：

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
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

// 执行天气工具
func (w *WeatherTool) Execute(args map[string]interface{}) (interface{}, error) {
	// 这里简化实现，返回模拟数据
	return map[string]interface{}{
		"temperature": 23.5,
		"condition":   "晴朗",
		"humidity":    65,
	}, nil
}

// 工具工厂，用于管理和注册工具
type ToolFactory struct {
	tools map[string]Tool
}

// 创建新的工具工厂
func NewToolFactory() *ToolFactory {
	return &ToolFactory{
		tools: make(map[string]Tool),
	}
}

// 注册工具
func (f *ToolFactory) RegisterTool(name string, tool Tool) {
	f.tools[name] = tool
}

// 获取工具定义列表
func (f *ToolFactory) GetToolDefinitions() []llms.Tool {
	definitions := make([]llms.Tool, 0, len(f.tools))
	for _, tool := range f.tools {
		definitions = append(definitions, tool.GetDefinition())
	}
	return definitions
}

// 执行工具
func (f *ToolFactory) ExecuteTool(name string, args map[string]interface{}) (interface{}, error) {
	tool, exists := f.tools[name]
	if !exists {
		return nil, fmt.Errorf("工具未找到: %s", name)
	}
	return tool.Execute(args)
}

func main() {
	// 初始化DeepSeek客户端
	llm, err := deepseek.New(
		deepseek.WithAPIKey("your-api-key"),
	)
	if err != nil {
		log.Fatal("初始化DeepSeek客户端失败: ", err)
	}

	ctx := context.Background()

	// 创建工具工厂
	toolFactory := NewToolFactory()

	// 注册天气工具
	toolFactory.RegisterTool("get_weather", &WeatherTool{})

	// 创建聊天消息
	systemPrompt := "你是一个旅行助手，可以帮助用户查询天气信息并提供相应的旅行建议。"
	userPrompt := "我想知道北京明天的天气如何，我应该准备什么衣物？"

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
	}

	// 生成内容并处理工具调用
	completion, err := llm.GenerateContent(
		ctx,
		content,
		llms.WithMaxTokens(1000),
		llms.WithTemperature(0.7),
		llms.WithTools(toolFactory.GetToolDefinitions()),
	)
	if err != nil {
		log.Fatal("生成内容失败: ", err)
	}

	// 处理工具调用
	if len(completion.Choices) > 0 && len(completion.Choices[0].ToolCalls) > 0 {
		// 获取第一个工具调用
		toolCall := completion.Choices[0].ToolCalls[0]
		fmt.Printf("模型请求调用工具: %s\n", toolCall.FunctionCall.Name)
		fmt.Printf("参数: %s\n", toolCall.FunctionCall.Arguments)

		// 解析参数
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
			log.Fatal("解析参数失败: ", err)
		}

		// 执行工具
		result, err := toolFactory.ExecuteTool(toolCall.FunctionCall.Name, args)
		if err != nil {
			log.Fatal("执行工具失败: ", err)
		}

		// 将结果转换为JSON
		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Fatal("序列化结果失败: ", err)
		}

		fmt.Printf("工具返回结果: %s\n\n", string(resultJSON))

		// 将工具调用结果发送回模型
		toolResult := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
			llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
			// 助手消息包含工具调用
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					// 添加工具调用
					llms.ToolCall{
						ID:   toolCall.ID,
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      toolCall.FunctionCall.Name,
							Arguments: toolCall.FunctionCall.Arguments,
						},
					},
				},
			},
			// 工具响应消息
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						Content:    string(resultJSON),
						ToolCallID: toolCall.ID,
						Name:       toolCall.FunctionCall.Name,
					},
				},
			},
		}

		// 获取最终回复
		finalResponse, err := llm.GenerateContent(ctx, toolResult, llms.WithMaxTokens(1000))
		if err != nil {
			log.Fatal("获取最终回复失败: ", err)
		}

		// 输出最终回复
		if len(finalResponse.Choices) > 0 {
			fmt.Println("最终回复:")
			fmt.Println(finalResponse.Choices[0].Content)
		}
	} else {
		// 直接返回回复
		if len(completion.Choices) > 0 {
			fmt.Println("模型回复:")
			fmt.Println(completion.Choices[0].Content)
		}
	}
}
```

## 注意事项与最佳实践

### 工具调用（Function Calling）

- 工具定义应尽可能详细，包括清晰的描述和参数说明，这有助于模型正确理解和使用工具。
- 处理工具调用时，确保正确处理错误情况，如参数解析错误、工具执行错误等。
- 在将工具调用结果发送回模型时，确保消息格式正确，特别是助手消息的格式。
- 对于复杂的工具调用场景，可以实现多轮工具调用，即模型可以根据前一个工具的结果决定是否调用下一个工具。

### API密钥管理

- 推荐使用环境变量管理API密钥，避免硬编码在代码中。
- 可以设置`DEEPSEEK_API_KEY`环境变量，库会自动读取。

```bash
export DEEPSEEK_API_KEY="your-api-key"
```

## 示例代码

完整的示例代码可以在[examples](./examples)目录中找到：

- [基本调用示例](./examples/deepseek-completion-example)
- [流式输出示例](./examples/deepseek-streaming-example)
- [工具调用示例](./examples/deepseek-tool-call-example)
- [多模态输入示例](./examples/deepseek-multimodal-example)

## 许可证

MIT