# LangChain-Go 中国版示例集合

本目录包含了 LangChain-Go 中国版的各种使用示例，展示如何使用国产大语言模型（如智谱AI、DeepSeek、通义千问、Kimi等）进行各种AI应用开发。

## 🚀 快速开始

### 环境准备

在运行任何示例之前，请确保设置了相应的API密钥环境变量：

```bash
# 智谱AI
export ZHIPU_API_KEY="your-zhipu-api-key"

# DeepSeek
export DEEPSEEK_API_KEY="your-deepseek-api-key"

# 通义千问
export QWEN_API_KEY="your-qwen-api-key"

# Kimi (月之暗面)
export KIMI_API_KEY="your-kimi-api-key"

# 硅基流动
export SILICONFLOW_API_KEY="your-siliconflow-api-key"

# OpenAI (可选，用于对比测试)
export OPENAI_API_KEY="your-openai-api-key"
```

### 支持的模型

| 厂商 | 模型标识 | 环境变量 | 说明 |
|------|---------|----------|------|
| 智谱AI | `Zhipu` | `ZHIPU_API_KEY` | GLM-4、GLM-4V等模型 |
| DeepSeek | `DeepSeek` | `DEEPSEEK_API_KEY` | DeepSeek-Chat、DeepSeek-Vision等 |
| 通义千问 | `Qwen` | `QWEN_API_KEY` | Qwen-Max、Qwen-VL等模型 |
| Kimi | `Kimi` | `KIMI_API_KEY` | Moonshot系列模型 |
| 硅基流动 | `SiliconFlow` | `SILICONFLOW_API_KEY` | 多种开源模型集合平台 |

## 📚 示例目录

### 1. 模型列表示例 (`model-list/`)

**功能**: 展示各个LLM支持的模型列表

**使用方法**:
```bash
cd model-list

# 查看所有支持的模型
go run main.go
```

**示例功能**:
- 列出智谱AI支持的所有模型
- 列出DeepSeek支持的所有模型  
- 列出通义千问支持的所有模型
- 列出Kimi支持的所有模型
- 列出硅基流动支持的文本生成、多模态和Embedding模型
- 展示如何使用GetModels()方法

**支持的LLM**:
- 智谱AI: GLM-4、GLM-4V、GLM-4-Air等8个模型
- DeepSeek: deepseek-chat、deepseek-coder等4个模型
- 通义千问: qwen-turbo、qwen-plus、qwen-max等5个模型
- Kimi: moonshot-v1-8k、moonshot-v1-32k、moonshot-v1-128k
- 硅基流动: 16个文本生成模型 + 3个多模态模型 + 4个Embedding模型

### 2. 文本补全示例 (`completion/`)

**功能**: 展示基础的文本生成和对话功能

**使用方法**:
```bash
cd completion

# 测试所有模型
go run main.go

# 测试特定模型
go run main.go Zhipu     # 智谱AI
go run main.go DeepSeek     # DeepSeek  
go run main.go Qwen        # 通义千问
go run main.go Kimi        # Kimi
go run main.go SiliconFlow # 硅基流动
```

**示例功能**:
- 简单的文本问答
- 对比不同模型的回答
- 展示基本的参数配置（温度、最大token数等）

---

### 2. 流式输出示例 (`streaming/`)

**功能**: 展示实时流式文本生成，适用于聊天场景

**使用方法**:
```bash
cd streaming

# 测试所有模型的流式输出
go run main.go

# 测试特定模型的流式输出
go run main.go Zhipu
go run main.go SiliconFlow
```

**示例功能**:
- 实时流式文本生成
- 逐字符或逐词输出
- 适合构建聊天界面

**技术特点**:
- 使用 `WithStreamingFunc` 回调函数
- 实时显示生成过程
- 低延迟用户体验

---

### 3. 向量嵌入示例 (`embedding/`)

**功能**: 展示文本向量化功能，用于语义搜索、相似度计算等

**使用方法**:
```bash
cd embedding

# 测试所有支持embedding的模型
go run main.go

# 测试特定模型
go run main.go Qwen        # 通义千问embedding
go run main.go Zhipu       # 智谱AI embedding
go run main.go SiliconFlow # 硅基流动embedding
```

**示例功能**:
- 文本向量化
- 计算文本相似度
- 支持批量处理

**应用场景**:
- 语义搜索
- 文档聚类
- 推荐系统
- RAG (检索增强生成)

---

### 4. 多模态示例 (`multi-modal/`)

**功能**: 展示图像理解和视觉推理能力

**使用方法**:
```bash
cd multi-modal

# 测试所有支持视觉的模型
go run main.go

# 测试特定模型
go run main.go Zhipu       # GLM-4V
go run main.go Qwen        # Qwen-VL
go run main.go Kimi        # Moonshot-Vision
go run main.go SiliconFlow # Qwen2-VL等
```

**示例功能**:
- 图像内容分析
- 图文对话
- 视觉问答

**支持的模型**:
- 智谱AI: GLM-4V
- 通义千问: Qwen-VL-Max
- Kimi: Moonshot-Vision
- DeepSeek: DeepSeek-Vision
- 硅基流动: Qwen2-VL、InternVL2等

---

### 5. 工具调用示例 (`toolcall/`)

**功能**: 展示函数调用(Function Calling)能力，让AI调用外部工具

**使用方法**:
```bash
cd toolcall

# 测试所有支持工具调用的模型
go run main.go

# 测试特定模型
go run main.go Zhipu
```

**示例功能**:
- 天气查询工具
- 计算器工具
- 自定义函数调用
- 工具链式调用

**内置工具**:
- `get_weather`: 查询天气信息
- `calculate`: 执行数学计算
- `get_time`: 获取当前时间

**技术特点**:
- JSON Schema 定义工具
- 参数验证
- 错误处理
- 工具结果反馈

## 🔧 高级用法

### 模型参数配置

所有示例都支持以下通用参数配置：

```go
response, err := llm.GenerateContent(
    ctx,
    messages,
    llms.WithTemperature(0.7),      // 创造性控制 (0.0-1.0)
    llms.WithMaxTokens(1000),       // 最大输出token数
    llms.WithTopP(0.9),             // 核采样参数
    llms.WithPresencePenalty(0.1),  // 存在惩罚
    llms.WithFrequencyPenalty(0.1), // 频率惩罚
)
```

### 错误处理

示例中包含了完整的错误处理机制：

```go
if err != nil {
    // 检查是否是API密钥问题
    if strings.Contains(err.Error(), "API密钥") {
        fmt.Println("请检查API密钥设置")
        return
    }
    
    // 检查是否是网络问题
    if strings.Contains(err.Error(), "timeout") {
        fmt.Println("请检查网络连接")
        return
    }
    
    log.Printf("未知错误: %v", err)
}
```

### 自定义配置

每个模型都支持自定义配置：

```go
// 智谱AI自定义配置
zhipuLLM, err := zhipu.New(
    zhipu.WithAPIKey("your-api-key"),
    zhipu.WithModel(zhipu.ModelGLM4V),
    zhipu.WithBaseURL("https://open.bigmodel.cn/api/paas/v4/"),
)

// DeepSeek自定义配置
deepseekLLM, err := deepseek.New(
    deepseek.WithAPIKey("your-api-key"),
    deepseek.WithModel("deepseek-chat"),
    deepseek.WithBaseURL("https://api.deepseek.com"),
)
```

## 🛠️ 开发指南

### 添加新示例

1. 在 `examples/` 目录下创建新文件夹
2. 添加 `main.go` 和 `go.mod` 文件
3. 使用统一的模型初始化模式：

```go
import cnllms "github.com/sjzsdu/langchaingo-cn/llms"

// 初始化文本模型
models, modelNames, err := cnllms.InitTextModels(llm)

// 初始化多模态模型  
models, modelNames, err := cnllms.InitImageModels(llm)

// 初始化embedding模型
models, modelNames, err := cnllms.InitEmbeddingModels(llm)
```

### 调试技巧

1. **启用详细日志**:
```bash
export DEBUG=true
go run main.go
```

2. **测试单个模型**:
```bash
go run main.go Zhipu
```

3. **检查API连接**:
```bash
curl -H "Authorization: Bearer $ZHIPU_API_KEY" \
     https://open.bigmodel.cn/api/paas/v4/chat/completions
```

## 📝 常见问题

### Q: 为什么某些模型不支持system消息？

A: 部分国产模型（如智谱AI）不支持OpenAI的system角色。我们的实现会自动将system消息转换为user消息，确保兼容性。

### Q: 如何选择合适的模型？

A: 
- **日常对话**: Kimi、智谱GLM-4
- **代码生成**: DeepSeek-Coder 
- **视觉理解**: GLM-4V、Qwen-VL
- **数学推理**: DeepSeek-Math
- **长文本处理**: Kimi (支持200K+ context)

### Q: 流式输出延迟较高怎么办？

A: 
1. 检查网络连接
2. 尝试降低 `max_tokens` 参数
3. 使用更快的模型（如GLM-4-Flash）

### Q: embedding向量维度是多少？

A: 
- 智谱AI embedding-2: 1024维
- 通义千问 text-embedding-v1: 1536维
- 各模型维度可能不同，请参考官方文档

## 🔗 相关链接

- [智谱AI开放平台](https://open.bigmodel.cn/)
- [DeepSeek API文档](https://platform.deepseek.com/)  
- [通义千问API文档](https://help.aliyun.com/zh/dashscope/)
- [Kimi API文档](https://platform.moonshot.cn/)
- [LangChain-Go 官方文档](https://github.com/tmc/langchaingo)

## 📄 许可证

本项目遵循 MIT 许可证。详情请参见项目根目录下的 LICENSE 文件。