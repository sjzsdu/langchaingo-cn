# LLM 工厂模式示例

本示例展示了如何使用工厂模式创建和使用不同类型的大语言模型（LLM）。

## 功能特性

- 支持通过统一的工厂方法创建不同类型的LLM
- 支持DeepSeek、Kimi、通义千问（Qwen）、Anthropic、OpenAI和Ollama六种LLM
- 支持通过参数映射配置LLM的各种选项
- 提供统一的接口进行文本生成和多轮对话

## 工厂模式设计

工厂模式是一种创建型设计模式，它提供了一种创建对象的最佳方式。在这个示例中，我们使用工厂模式来创建不同类型的LLM，主要优势包括：

1. **抽象创建过程**：客户端代码不需要了解具体LLM的创建细节
2. **统一接口**：所有LLM通过相同的接口进行交互
3. **易于扩展**：添加新的LLM类型只需扩展工厂方法，不需要修改客户端代码
4. **参数灵活性**：通过参数映射支持不同LLM的特定配置选项

## 使用方法

### 环境准备

1. 设置API密钥环境变量（根据需要设置）：

```bash
# 用于DeepSeek、Kimi和Qwen
export LLM_API_KEY=your-api-key

# 用于Anthropic
export ANTHROPIC_API_KEY=your-anthropic-api-key

# 用于OpenAI
export OPENAI_API_KEY=your-openai-api-key

# Ollama不需要API密钥，但需要在本地运行Ollama服务
# 默认URL为http://localhost:11434
```

2. 运行示例：

```bash
go run main.go
```

### 代码示例

```go
// 创建DeepSeek LLM
deepseekLLM, err := llmscn.CreateLLM(llmscn.DeepSeekLLM, map[string]interface{}{
    "api_key": apiKey,
    "model":   "deepseek-chat",
})

// 创建Kimi LLM
kimiLLM, err := llmscn.CreateLLM(llmscn.KimiLLM, map[string]interface{}{
    "api_key":      apiKey,
    "model":        "moonshot-v1-8k",
    "temperature":  0.7,
    "top_p":        0.9,
    "max_tokens":   1000,
})

// 创建Qwen LLM
qwenLLM, err := llmscn.CreateLLM(llmscn.QwenLLM, map[string]interface{}{
    "api_key":     apiKey,
    "model":       "qwen-turbo",
    "temperature": 0.8,
    "top_p":       0.95,
    "top_k":       50,
    "max_tokens":  2000,
})

// 创建Anthropic LLM
anthropicLLM, err := llmscn.CreateLLM(llmscn.AnthropicLLM, map[string]interface{}{
    "api_key": anthropicAPIKey,
    "model":   "claude-3-sonnet-20240229",
})

// 创建OpenAI LLM
openaiLLM, err := llmscn.CreateLLM(llmscn.OpenAILLM, map[string]interface{}{
    "api_key":      openaiAPIKey,
    "model":        "gpt-4",
    "organization": "your-org-id",  // 可选
    "api_type":     "openai",      // 可选，可选值："openai"、"azure"、"azure_ad"
})

// 创建Ollama LLM
ollamaLLM, err := llmscn.CreateLLM(llmscn.OllamaLLM, map[string]interface{}{
    "server_url": "http://localhost:11434",  // 可选，默认为"http://localhost:11434"
    "model":      "llama3",
    "format":     "json",                   // 可选
    "system":     "你是一个有用的AI助手。",     // 可选
})
```

## 支持的参数

### 通用参数

所有LLM类型都支持以下参数：

- `api_key`：API密钥（必需）
- `model`：模型名称
- `base_url`：API基础URL
- `temperature`：温度参数
- `top_p`：Top-P参数
- `max_tokens`：最大生成令牌数

### 特定参数

- **Qwen**：
  - `top_k`：Top-K参数
  - `use_openai_compatible`：是否使用OpenAI兼容模式

- **OpenAI**：
  - `organization`：组织ID
  - `api_type`：API类型（可选值："openai"、"azure"、"azure_ad"）
  - `api_version`：API版本（默认为"2023-05-15"）

- **Ollama**：
  - `server_url`：Ollama服务器URL（默认为"http://localhost:11434"）
  - `format`：输出格式（可选值："json"）
  - `system`：系统提示

## 注意事项

- 确保已安装所需的依赖包
- 不同的LLM可能需要不同的API密钥
- 部分高级功能（如多模态内容处理）需要通过特定LLM的原生接口实现
- 使用Ollama需要在本地运行Ollama服务，可以从[Ollama官网](https://ollama.ai/)下载安装
- 使用Anthropic和OpenAI需要有效的API密钥，可以从各自的官方网站获取
- 不同的LLM可能支持不同的模型，请根据实际情况选择合适的模型名称