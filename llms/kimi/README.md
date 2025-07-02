# Kimi LLM

本包提供了与 Moonshot AI 的 Kimi 大语言模型交互的功能。

## 功能特性

- 支持基本文本生成
- 支持流式响应
- 支持多模态内容处理
- 支持工具调用

## 安装

```bash
go get github.com/sjzsdu/langchaingo-cn
```

## 使用方法

### 初始化客户端

```go
import (
    "github.com/sjzsdu/langchaingo-cn/llms/kimi"
)

// 创建Kimi LLM客户端
llm, err := kimi.New(
    kimi.WithToken("your-api-key"),
    kimi.WithModel(kimi.ModelKimi),
    kimi.WithTemperature(0.7),
)
```

### 基本调用

```go
resp, err := llm.Call(context.Background(), "你好，请介绍一下自己")
if err != nil {
    // 处理错误
}
fmt.Println(resp)
```

### 流式调用

```go
stream, err := llm.StreamingCall(context.Background(), "请解释量子计算的基本原理")
if err != nil {
    // 处理错误
}

for {
    chunk, err := stream.GetChunk()
    if err != nil {
        break
    }
    fmt.Print(chunk)
}
```

### 多模态内容处理

```go
content := []kimi.ContentPart{
    kimi.ContentPartText("请分析下面的代码："),
    kimi.ContentPartText("```go\nfunc fibonacci(n int) int {\n\tif n <= 1 {\n\t\treturn n\n\t}\n\treturn fibonacci(n-1) + fibonacci(n-2)\n}\n```"),
}

contentResp, err := llm.GenerateContent(context.Background(), content, nil)
if err != nil {
    // 处理错误
}

fmt.Println(contentResp.Content)
```

## 配置选项

- `WithToken(token string)`：设置 API 密钥
- `WithModel(model string)`：选择模型，可用值：
  - `kimi.ModelKimi`：Kimi 基础模型
  - `kimi.ModelKimiPro`：Kimi Pro 高级模型
- `WithTemperature(temperature float64)`：设置温度参数（0-1之间）
- `WithMaxTokens(maxTokens int)`：设置最大生成令牌数
- `WithTopP(topP float64)`：设置 top_p 参数
- `WithBaseURL(baseURL string)`：自定义 API 基础 URL

## 环境变量

- `MOONSHOT_API_KEY`：Moonshot AI API 密钥

## 错误处理

```go
resp, err := llm.Call(context.Background(), "你好")
if err != nil {
    if errors.Is(err, kimi.ErrEmptyResponse) {
        // 处理空响应错误
    } else if errors.Is(err, kimi.ErrRequestFailed) {
        // 处理请求失败错误
    } else {
        // 处理其他错误
    }
}
```

## 示例

完整示例请参考 [examples/kimi-example](../../examples/kimi-example) 目录。