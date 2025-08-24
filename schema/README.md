# Schema 包 - 配置驱动的组件工厂系统

Schema 包为 LangChainGo-CN 提供了一个强大的配置驱动组件工厂系统，允许用户通过 JSON 配置文件来定义和创建各种 LangChain 组件。

## 功能特性

- ✅ **配置驱动**: 通过 JSON 配置文件定义组件
- ✅ **组件工厂**: 支持 LLM、Memory、Prompt、Embedding、Chain、Agent 等组件
- ✅ **环境变量支持**: 自动展开 `${VARIABLE}` 格式的环境变量
- ✅ **依赖解析**: 自动处理组件间的引用关系
- ✅ **配置验证**: 完整的配置验证和错误报告
- ✅ **类型安全**: 确保创建的组件符合相应接口

## 支持的组件类型

### LLM 组件
- `openai`: OpenAI GPT 模型
- `deepseek`: DeepSeek 模型
- `kimi`: Kimi 月之暗面模型
- `qwen`: 通义千问模型
- `anthropic`: Anthropic Claude 模型
- `ollama`: 本地 Ollama 模型

### Memory 组件
- `conversation_buffer`: 会话缓冲记忆
- `conversation_summary`: 会话摘要记忆
- `conversation_token_buffer`: 基于 Token 的会话记忆
- `simple`: 简单记忆

### Prompt 组件
- `prompt_template`: 普通提示模板
- `chat_prompt_template`: 聊天提示模板

### Embedding 组件
- `openai`: OpenAI 嵌入模型
- `voyage`: VoyageAI 嵌入模型
- `cohere`: Cohere 嵌入模型

### Chain 组件
- `llm`: 基础 LLM 链
- `conversation`: 对话链
- `sequential`: 顺序链
- `stuff_documents`: 文档填充链
- `map_reduce`: MapReduce 链

### Agent 组件
- `zero_shot_react`: 零样本 ReAct 智能体
- `conversational_react`: 对话式 ReAct 智能体

## 快速开始

### 1. 基本用法

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/sjzsdu/langchaingo-cn/schema"
)

func main() {
    // 从配置文件创建应用
    app, err := schema.CreateApplicationFromFile("config.json")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("创建了 %d 个组件\n", len(app.LLMs))
}
```

### 2. JSON 配置示例

```json
{
  "llms": {
    "main_llm": {
      "type": "deepseek",
      "model": "deepseek-chat",
      "api_key": "${DEEPSEEK_API_KEY}",
      "temperature": 0.7
    }
  },
  "memories": {
    "chat_memory": {
      "type": "conversation_buffer",
      "max_messages": 10
    }
  },
  "chains": {
    "chat_chain": {
      "type": "conversation",
      "llm_ref": "main_llm",
      "memory_ref": "chat_memory"
    }
  }
}
```

### 3. 使用创建的组件

```go
// 获取创建的链并使用
if chain, exists := app.Chains["chat_chain"]; exists {
    result, err := chains.Run(context.Background(), chain, "你好")
    if err != nil {
        log.Printf("执行失败: %v", err)
    } else {
        fmt.Printf("AI回复: %s\n", result)
    }
}
```

## 详细配置说明

### LLM 配置

```json
{
  "type": "openai",           // 必需：LLM 类型
  "model": "gpt-4",          // 必需：模型名称
  "api_key": "${API_KEY}",   // API 密钥（支持环境变量）
  "base_url": "https://...", // 可选：自定义 API 基础 URL
  "temperature": 0.7,        // 可选：温度参数
  "max_tokens": 2048,        // 可选：最大 Token 数
  "options": {               // 可选：其他选项
    "organization": "org-id"
  }
}
```

### Memory 配置

```json
{
  "type": "conversation_summary",  // 必需：记忆类型
  "llm_ref": "summary_llm",       // 可选：引用的 LLM（某些类型需要）
  "max_token_limit": 1000,        // 可选：Token 限制
  "max_messages": 10,             // 可选：消息数量限制
  "return_messages": true         // 可选：是否返回消息
}
```

### Chain 配置

```json
{
  "type": "conversation",    // 必需：链类型
  "llm_ref": "main_llm",    // 可选：引用的 LLM
  "memory_ref": "memory",   // 可选：引用的 Memory
  "prompt_ref": "prompt",   // 可选：引用的 Prompt
  "chains": ["chain1"],     // 可选：子链（用于 sequential）
  "input_keys": ["input"],  // 可选：输入键
  "output_keys": ["output"] // 可选：输出键
}
```

## 环境变量

设置相应的环境变量来提供 API 密钥：

```bash
export OPENAI_API_KEY="your-openai-key"
export DEEPSEEK_API_KEY="your-deepseek-key"
export KIMI_API_KEY="your-kimi-key"
export QWEN_API_KEY="your-qwen-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
```

## 配置验证

Schema 包提供了完整的配置验证功能：

```go
// 验证配置
config, err := schema.LoadConfigFromFile("config.json")
if err != nil {
    log.Fatal(err)
}

result := schema.ValidateConfig(config)
if result.HasErrors() {
    fmt.Printf("配置错误:\n%s\n", result.String())
    return
}

if result.HasWarnings() {
    fmt.Printf("配置警告:\n%s\n", result.String())
}
```

## 错误处理

Schema 包提供了结构化的错误类型：

```go
app, err := schema.CreateApplicationFromFile("config.json")
if err != nil {
    if schemaErr, ok := err.(*schema.SchemaError); ok {
        fmt.Printf("错误类型: %s\n", schemaErr.Type)
        fmt.Printf("错误路径: %s\n", schemaErr.Path)
        fmt.Printf("错误消息: %s\n", schemaErr.Message)
    }
}
```

## 示例

### 简单聊天应用

查看 `examples/simple_chat.json` 了解如何配置一个基本的聊天应用。

### 复杂应用

查看 `examples/complex_app.json` 了解如何配置包含多个组件类型的复杂应用。

### 完整用法示例

运行 `examples/usage_example.go` 查看完整的使用示例。

## API 参考

### 主要函数

- `CreateApplicationFromFile(filename string) (*Application, error)`: 从文件创建应用
- `CreateApplicationFromJSON(jsonStr string) (*Application, error)`: 从 JSON 字符串创建应用
- `LoadConfigFromFile(filename string) (*Config, error)`: 从文件加载配置
- `LoadConfigFromJSON(jsonStr string) (*Config, error)`: 从 JSON 加载配置
- `ValidateConfig(config *Config) *ValidationResult`: 验证配置

### 单组件创建函数

- `CreateLLMFromConfig(config *LLMConfig) (llms.Model, error)`
- `CreateMemoryFromConfig(config *MemoryConfig, llmConfigs map[string]*LLMConfig) (schema.Memory, error)`
- `CreatePromptFromConfig(config *PromptConfig) (prompts.PromptTemplate, error)`
- `CreateEmbeddingFromConfig(config *EmbeddingConfig) (embeddings.Embedder, error)`

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个包。请确保：

1. 添加适当的测试用例
2. 更新文档
3. 遵循现有的代码风格

## 许可证

本项目采用与主项目相同的许可证。