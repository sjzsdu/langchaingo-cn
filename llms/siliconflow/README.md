# 硅基流动 (SiliconFlow) LLM

硅基流动大语言模型的Go语言实现，基于OpenAI兼容API接口。

## 功能特性

- 支持硅基流动平台的多个开源模型，包括Qwen、DeepSeek、GLM等
- 完全兼容OpenAI API格式，方便迁移
- 支持文本生成、对话、embedding、多模态等功能
- 支持流式响应
- 高性价比，部分模型免费使用
- 完整的配置选项支持

## 安装

```bash
go get github.com/tmc/langchaingo/llms/siliconflow
```

## 快速开始

### 1. 获取API Key

访问 [硅基流动平台](https://cloud.siliconflow.cn/account/ak) 注册账号并获取API Key。

### 2. 设置环境变量

```bash
export SILICONFLOW_API_KEY="your-api-key-here"
export SILICONFLOW_MODEL="Qwen/Qwen2.5-72B-Instruct"  # 可选，默认为Qwen2.5-72B
```

### 3. 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/tmc/langchaingo/llms/siliconflow"
    "github.com/tmc/langchaingo/schema"
)

func main() {
    // 创建LLM实例
    llm, err := siliconflow.New()
    if err != nil {
        log.Fatal(err)
    }

    // 生成文本
    response, err := llm.Call(context.Background(), "你好，请介绍一下硅基流动平台")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response)
}
```

### 4. 使用配置选项

```go
llm, err := siliconflow.New(
    siliconflow.WithAPIKey("your-api-key"),
    siliconflow.WithModel(siliconflow.ModelQwen2572B),
    siliconflow.WithBaseURL("https://api.siliconflow.cn/v1"),
)
```

### 5. 多轮对话

```go
messages := []schema.ChatMessage{
    schema.SystemChatMessage{Content: "你是一个有用的助手"},
    schema.HumanChatMessage{Content: "硅基流动有什么优势？"},
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
    schema.HumanChatMessage{Content: "写一篇关于AI发展的文章"},
}

err := llm.GenerateContentStream(context.Background(), messages, 
    func(ctx context.Context, chunk []byte) error {
        fmt.Print(string(chunk))
        return nil
    })
```

### 7. 使用推理模型

```go
// 使用DeepSeek-R1推理模型
llm, err := siliconflow.New(
    siliconflow.WithModel(siliconflow.ModelDeepSeekR1),
)

// 推理模型会输出思维过程
response, err := llm.GenerateContent(context.Background(), messages)
```

## 支持的模型

### 文本生成模型

| 模型名称 | 常量 | 描述 | 价格 |
|---------|------|------|------|
| Qwen/Qwen2.5-72B-Instruct | `siliconflow.ModelQwen2572B` | 通义千问2.5-72B，性能强劲 | 付费 |
| Qwen/Qwen2.5-7B-Instruct | `siliconflow.ModelQwen257B` | 通义千问2.5-7B，免费使用 | **免费** |
| deepseek-ai/DeepSeek-V3 | `siliconflow.ModelDeepSeekV3` | DeepSeek最新模型 | 付费 |
| Pro/deepseek-ai/DeepSeek-R1 | `siliconflow.ModelDeepSeekR1` | DeepSeek推理模型 | 付费 |
| deepseek-ai/DeepSeek-V2.5 | `siliconflow.ModelDeepSeekV25` | DeepSeek-V2.5 | 付费 |
| ZHIPU/GLM-4-9B-Chat | `siliconflow.ModelGLM49B` | 智谱GLM-4-9B | 付费 |
| Qwen/QwQ-32B-Preview | `siliconflow.ModelQwQ32B` | QwQ推理模型 | 付费 |

### 多模态模型

| 模型名称 | 常量 | 描述 |
|---------|------|------|
| Qwen/Qwen2-VL-72B-Instruct | `siliconflow.ModelQwenVLMax` | 通义千问视觉72B模型 |
| Qwen/Qwen2-VL-7B-Instruct | `siliconflow.ModelQwenVL7B` | 通义千问视觉7B模型 |
| OpenGVLab/InternVL2-26B | `siliconflow.ModelInternVL2` | InternVL2多模态模型 |

### Embedding模型

| 模型名称 | 常量 | 描述 | 维度 |
|---------|------|------|------|
| BAAI/bge-large-zh-v1.5 | `siliconflow.ModelBGELargeZh` | BGE中文向量模型 | 1024 |
| BAAI/bge-base-zh-v1.5 | `siliconflow.ModelBGEBaseZh` | BGE基础中文向量模型 | 768 |
| maidalun1020/bce-embedding-base_v1 | `siliconflow.ModelBCEEmbedding` | BCE向量模型 | 768 |

## 配置选项

### 环境变量

- `SILICONFLOW_API_KEY`: 硅基流动API密钥（必需）
- `SILICONFLOW_MODEL`: 默认使用的模型（可选，默认为Qwen/Qwen2.5-72B-Instruct）
- `SILICONFLOW_EMBEDDING_MODEL`: Embedding模型（可选，默认为BAAI/bge-large-zh-v1.5）

### 函数选项

- `WithAPIKey(string)`: 设置API密钥
- `WithModel(string)`: 设置默认模型
- `WithBaseURL(string)`: 设置API基础URL
- `WithEmbeddingModel(string)`: 设置Embedding模型

## 多模态使用

```go
// 创建多模态LLM实例
llm, err := siliconflow.New(
    siliconflow.WithModel(siliconflow.ModelQwenVLMax),
)

// 图像分析
messages := []schema.ChatMessage{
    schema.HumanChatMessage{
        Content: []schema.ContentPart{
            schema.ImageURLPart{URL: "https://example.com/image.jpg"},
            schema.TextPart{Text: "这张图片里有什么？"},
        },
    },
}

response, err := llm.GenerateContent(context.Background(), messages)
```

## Embedding 使用

```go
// 创建embedding实例
llm, err := siliconflow.New(
    siliconflow.WithEmbeddingModel(siliconflow.ModelBGELargeZh),
)

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
llm, err := siliconflow.New()
if err != nil {
    switch {
    case strings.Contains(err.Error(), "API密钥不能为空"):
        fmt.Println("请设置SILICONFLOW_API_KEY环境变量")
    default:
        log.Fatal(err)
    }
}
```

## 特殊功能

### 推理模型

硅基流动支持推理模型（如DeepSeek-R1），这些模型会输出思维过程：

```go
llm, err := siliconflow.New(
    siliconflow.WithModel(siliconflow.ModelDeepSeekR1),
)

// 推理模型在生成答案前会展示思考过程
response, err := llm.GenerateContent(ctx, messages, 
    llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
        // 可以捕获推理过程和最终答案
        fmt.Print(string(chunk))
        return nil
    }))
```

### 免费模型

硅基流动提供多个免费模型，非常适合开发和测试：

```go
// 使用免费的Qwen2.5-7B模型
llm, err := siliconflow.New(
    siliconflow.WithModel(siliconflow.ModelQwen257B),
)
```

## 性能优势

- **高速推理**: 10x+ 速度提升
- **高性价比**: 相比直接调用原厂API节省46-66%成本
- **高稳定性**: 企业级服务保障
- **免费额度**: 多个模型提供免费使用

## 注意事项

1. **API密钥安全**: 不要在代码中硬编码API密钥，建议使用环境变量
2. **模型选择**: 免费模型适合开发测试，生产环境可选择性能更强的付费模型
3. **速率限制**: 不同模型有不同的调用频率限制，请合理控制调用频率
4. **推理模型**: 推理模型输出包含思维过程，token消耗会更多

## 相关链接

- [硅基流动官网](https://siliconflow.cn/)
- [硅基流动平台](https://cloud.siliconflow.cn/)
- [API文档](https://docs.siliconflow.cn/)
- [模型广场](https://cloud.siliconflow.cn/models)
- [价格说明](https://siliconflow.cn/pricing)

## 许可证

本项目遵循与langchaingo主项目相同的许可证。