# 智谱AI (ZhipuAI) LLM

智谱AI大语言模型的Go语言实现，基于OpenAI兼容API接口。

## 功能特性

- 支持智谱AI的多个模型，包括GLM-4、GLM-4V等
- 兼容OpenAI API格式，方便迁移
- 支持文本生成、对话、embedding等功能
- 支持流式响应
- 完整的配置选项支持

## 安装

```bash
go get github.com/tmc/langchaingo/llms/zhipu
```

## 快速开始

### 1. 获取API Key

访问 [智谱AI开放平台](https://open.bigmodel.cn/) 注册账号并获取API Key。

### 2. 设置环境变量

```bash
export ZHIPU_API_KEY="your-api-key-here"
export ZHIPU_MODEL="glm-4"  # 可选，默认为glm-4
```

### 3. 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms/zhipu"
    "github.com/tmc/langchaingo/schema"
)

func main() {
    // 创建LLM实例
    llm, err := zhipu.New()
    if err != nil {
        log.Fatal(err)
    }

    // 生成文本
    response, err := llm.Call(context.Background(), "你好，请介绍一下自己")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response)
}
```

### 4. 使用配置选项

```go
llm, err := zhipu.New(
    zhipu.WithAPIKey("your-api-key"),
    zhipu.WithModel(zhipu.ModelGLM4),
    zhipu.WithBaseURL("https://open.bigmodel.cn/api/paas/v4/"),
)
```

### 5. 多轮对话

```go
messages := []schema.ChatMessage{
    schema.SystemChatMessage{Content: "你是一个有用的助手"},
    schema.HumanChatMessage{Content: "你好"},
}

response, err := llm.GenerateContent(context.Background(), messages)
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Choices[0].Content)
```

### 6. 流式响应

```go
messages := []schema.ChatMessage{
    schema.HumanChatMessage{Content: "写一篇关于人工智能的文章"},
}

err := llm.GenerateContentStream(context.Background(), messages, 
    func(ctx context.Context, chunk []byte) error {
        fmt.Print(string(chunk))
        return nil
    })
```

## 支持的模型

| 模型名称 | 常量 | 描述 |
|---------|------|------|
| glm-4 | `zhipu.ModelGLM4` | 智谱GLM-4主力模型 |
| glm-4v | `zhipu.ModelGLM4V` | 智谱GLM-4V视觉模型 |
| glm-4-air | `zhipu.ModelGLM4Air` | 智谱GLM-4-Air轻量级模型 |
| glm-4-airx | `zhipu.ModelGLM4AirX` | 智谱GLM-4-AirX模型 |
| glm-4-flash | `zhipu.ModelGLM4Flash` | 智谱GLM-4-Flash快速模型 |
| glm-3-turbo | `zhipu.ModelGLM3Turbo` | 智谱GLM-3-Turbo模型 |
| charglm-3 | `zhipu.ModelCharGLM3` | 智谱CharGLM-3角色扮演模型 |
| cogview-3 | `zhipu.ModelCogView3` | 智谱CogView-3图像生成模型 |

## 配置选项

### 环境变量

- `ZHIPU_API_KEY`: 智谱AI API密钥（必需）
- `ZHIPU_MODEL`: 默认使用的模型（可选，默认为glm-4）
- `ZHIPU_EMBEDDING_MODEL`: Embedding模型（可选，默认为embedding-2）

### 函数选项

- `WithAPIKey(string)`: 设置API密钥
- `WithModel(string)`: 设置默认模型
- `WithBaseURL(string)`: 设置API基础URL
- `WithEmbeddingModel(string)`: 设置Embedding模型

## Embedding 使用

```go
// 创建embedding
embeddings, err := llm.CreateEmbedding(context.Background(), []string{
    "这是第一段文本",
    "这是第二段文本",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("得到 %d 个embedding向量\n", len(embeddings))
```

## 错误处理

```go
llm, err := zhipu.New()
if err != nil {
    switch {
    case strings.Contains(err.Error(), "API密钥不能为空"):
        fmt.Println("请设置ZHIPU_API_KEY环境变量")
    default:
        log.Fatal(err)
    }
}
```

## 注意事项

1. **API密钥安全**: 不要在代码中硬编码API密钥，建议使用环境变量
2. **速率限制**: 智谱AI有API调用频率限制，请合理控制调用频率
3. **模型选择**: 不同模型有不同的能力和定价，请根据需求选择合适的模型
4. **网络代理**: 如需使用代理，可以通过标准的Go HTTP客户端配置实现

## 许可证

本项目遵循与langchaingo主项目相同的许可证。