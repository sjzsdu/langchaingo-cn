# DeepSeek Completion 示例

这个示例展示了如何使用 LangChain Go 与 DeepSeek API 进行简单的文本生成。

## 功能

- 使用 DeepSeek 模型生成文本回复
- 演示如何设置基本参数（温度、最大令牌数等）
- 使用单一提示进行文本生成

## 使用方法

1. 确保你有 DeepSeek API 密钥
2. 确保存在环境变量： `DEEPSEEK_API_KEY` 为你的实际 API 密钥
3. 运行示例：

```bash
go run deepseek-completion-example.go
```

## 代码说明

这个示例使用 `GenerateFromSinglePrompt` 函数，这是一个简单的方法，用于从单个文本提示生成回复。它适用于简单的问答场景，不需要复杂的对话历史或系统提示。

示例中使用了以下参数：

- `WithTemperature(0.7)`: 控制生成文本的随机性
- `WithMaxTokens(1000)`: 限制生成文本的最大长度

## 注意事项

- 确保你的 API 密钥有足够的配额
- 根据需要调整模型名称和参数