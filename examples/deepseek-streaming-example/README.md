# DeepSeek 流式输出示例

这个示例展示了如何使用 LangChain Go 与 DeepSeek API 进行流式文本生成，实时接收生成的内容。

## 功能

- 使用 DeepSeek 模型进行流式文本生成
- 实时接收和显示生成的内容
- 演示如何设置系统提示和用户消息

## 使用方法

1. 确保你有 DeepSeek API 密钥
2. 在代码中替换 `your-api-key` 为你的实际 API 密钥
3. 运行示例：

```bash
go run deepseek-streaming-example.go
```

## 代码说明

这个示例使用 `GenerateContent` 函数结合 `WithStreamingFunc` 选项来实现流式输出。每当模型生成新的内容块时，提供的回调函数就会被调用，使应用能够实时处理和显示生成的内容。

示例中使用了以下参数：

- `WithMaxTokens(1000)`: 限制生成文本的最大长度
- `WithTemperature(0.7)`: 控制生成文本的随机性
- `WithStreamingFunc`: 提供一个回调函数来处理流式输出

## 流式输出的优势

- 提供更好的用户体验，用户不需要等待整个响应完成
- 适用于长回答场景，可以立即开始显示内容
- 对于需要实时反馈的应用（如聊天机器人）非常有用

## 注意事项

- 流式输出会建立一个长连接，确保你的网络环境稳定
- 根据需要调整模型名称和参数