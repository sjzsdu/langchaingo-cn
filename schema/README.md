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
- `zhipu`: 智谱AI GLM 模型 🆕
- `siliconflow`: 硅基流动平台模型 🆕
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

## 模型列表查询

所有LLM实现都提供了 `GetModels()` 方法来枚举支持的模型列表：

```go
import (
    "fmt"
    "github.com/tmc/langchaingo-cn/llms/zhipu"
    "github.com/tmc/langchaingo-cn/llms/siliconflow"
)

// 获取智谱AI支持的模型
zhipuLLM, _ := zhipu.New(zhipu.WithAPIKey("your-key"))
models := zhipuLLM.GetModels()
fmt.Printf("智谱AI模型: %v\n", models)
// 输出: [glm-4 glm-4v glm-3-turbo]

// 获取硅基流动支持的模型
sfLLM, _ := siliconflow.New(siliconflow.WithAPIKey("your-key"))
models = sfLLM.GetModels()
fmt.Printf("硅基流动模型: %v\n", models)
// 输出: [Qwen/Qwen2-7B-Instruct deepseek-ai/DeepSeek-V2-Chat ...]
```

## 配置生成器 🚀

为了简化配置文件的创建，Schema 包提供了强大的配置生成器，可以快速生成各种组件的极简配置文件。

### 快速生成配置文件

```go
package main

import (
    "github.com/sjzsdu/langchaingo-cn/schema"
)

func main() {
    // 1. 快速生成LLM配置
    schema.QuickGenerateLLM("deepseek", "deepseek-chat", "my_llm.json")
    
    // 2. 快速生成Chain配置
    schema.QuickGenerateChain("conversation", "kimi", "moonshot-v1-8k", "my_chain.json")
    
    // 3. 快速生成Agent配置
    schema.QuickGenerateAgent("zero_shot_react", "openai", "gpt-4", "my_agent.json")
    
    // 4. 快速生成Executor配置
    schema.QuickGenerateExecutor("conversational_react", "qwen", "qwen-plus", "my_executor.json")
}
```

### 使用配置生成器

```go
// 创建配置生成器
generator := schema.NewConfigGenerator("./configs")

// 生成DeepSeek聊天配置
generator.GenerateDeepSeekChatConfig("deepseek_chat.json")

// 生成Kimi聊天配置
generator.GenerateKimiChatConfig("kimi_chat.json")

// 生成自定义Chain配置
generator.GenerateChainConfig(schema.ChainTemplate{
    Type: "conversation",
    LLMTemplate: schema.LLMTemplate{
        Type:        "deepseek",
        Model:       "deepseek-chat",
        Temperature: 0.7,
        MaxTokens:   2048,
    },
    MemoryType:     "conversation_buffer",
    PromptTemplate: "你是专业的AI助手，请回答：{{.input}}",
}, "custom_chain.json")
```

## 命令行工具

Schema 包提供了方便的命令行工具来快速生成配置文件：

### 基本用法

```bash
# 生成预设配置
go run main.go config-gen preset [preset-type] -o [output-file]

# 查看支持的命令和选项
go run main.go config-gen --help

# 列出所有可用的预设类型
go run main.go config-gen list
```

### 支持的预设类型

- `deepseek-chat`: DeepSeek 聊天配置
- `deepseek-executor`: DeepSeek 执行器配置
- `kimi-chat`: Kimi 聊天配置
- `openai-chat`: OpenAI 聊天配置
- `qwen-chat`: 通义千问聊天配置
- `zhipu-chat`: 智谱AI 聊天配置 🆕
- `zhipu-executor`: 智谱AI 执行器配置 🆕
- `siliconflow-chat`: 硅基流动 聊天配置 🆕
- `siliconflow-executor`: 硅基流动 执行器配置 🆕

### 配置验证 🆕

新增了配置文件验证命令，可以验证生成的JSON配置是否有效：

```bash
# 基础验证配置文件
go run main.go config-gen validate config.json

# 详细验证模式
go run main.go config-gen validate config.json --verbose

# 完整验证(包含真实API调用测试) 🚀
go run main.go config-gen validate config.json --api-test --verbose
```

该命令支持两种验证级别：

**基础验证** (默认):
- ✅ JSON语法和结构验证
- ✅ 组件类型和配置有效性检查  
- ✅ 实际组件创建测试
- ✅ GetModels()等基本功能测试
- ✅ 生成详细验证报告

**API测试验证** (`--api-test`):
- ✅ 包含所有基础验证功能
- 🌐 **真实API调用测试**: 发送测试请求给LLM
- 🌐 **Chain功能测试**: 验证对话链是否正常工作
- 🌐 **Agent执行测试**: 验证智能体是否能正常执行任务
- 📊 完整的功能验证报告

#### API测试工作原理 🚀

当使用 `--api-test` 参数时，验证器会执行以下真实测试：

**LLM测试**:
```
测试问题: "你是什么模型？请简短回答。"
验证标准: 收到非空响应且无错误
超时设置: 30秒
```

**Chain测试**:
```
测试输入: "你好，请告诉我你是什么AI助手？"
验证标准: 对话链正常执行并返回有效响应
超时设置: 30秒
```

**Agent测试**:
```
测试任务: "请简单介绍一下你自己的能力"
验证标准: 智能体成功执行并产生输出
超时设置: 45秒
```

⚠️ **注意**: API测试需要有效的API密钥和网络连接，测试会产生实际的API调用费用。

#### 验证最佳实践

**开发阶段建议**:
```bash
# 1. 首先进行基础验证，确保配置正确
go run main.go config-gen validate config.json --verbose

# 2. 配置无误后，进行API测试验证功能
go run main.go config-gen validate config.json --api-test --verbose

# 3. 生产环境部署前的最终验证
go run main.go config-gen validate production-config.json --api-test
```

**故障排除**:
- 如果基础验证失败：检查JSON语法和配置参数
- 如果API测试失败：验证API密钥设置和网络连接
- 如果超时：检查网络状况或增大超时设置

### 示例命令

```bash
# 生成智谱AI聊天配置
go run main.go config-gen preset zhipu-chat -o zhipu_config.json

# 生成硅基流动执行器配置
go run main.go config-gen preset siliconflow-executor -o sf_executor.json

# 生成DeepSeek聊天配置
go run main.go config-gen preset deepseek-chat -o deepseek_config.json

# 生成自定义LLM配置
go run main.go config-gen llm --llm zhipu --model glm-4 -o custom_zhipu.json

# 生成Chain配置
go run main.go config-gen chain --llm siliconflow --model Qwen/Qwen2-7B-Instruct -o sf_chain.json

# 验证生成的配置文件 🆕
go run main.go config-gen validate zhipu_config.json --verbose

# 完整API测试验证 🚀
go run main.go config-gen validate zhipu_config.json --api-test
```

### 预设配置快捷方法

```go
// DeepSeek相关
generator.GenerateDeepSeekChatConfig("deepseek_chat.json")
generator.GenerateExecutorWithDeepSeek("deepseek_executor.json")

// Kimi相关
generator.GenerateKimiChatConfig("kimi_chat.json")
generator.GenerateConversationalAgentConfig("kimi", "moonshot-v1-8k", "kimi_agent.json")

// OpenAI相关
generator.GenerateOpenAIChatConfig("openai_chat.json")
generator.GenerateReactAgentConfig("openai", "gpt-4", "openai_agent.json")

// 通义千问相关
generator.GenerateQwenChatConfig("qwen_chat.json")

// 智谱AI相关 🆕
generator.GenerateZhipuChatConfig("zhipu_chat.json")
generator.GenerateExecutorWithZhipu("zhipu_executor.json")

// 硅基流动相关 🆕
generator.GenerateSiliconFlowChatConfig("siliconflow_chat.json")
generator.GenerateExecutorWithSiliconFlow("siliconflow_executor.json")
```

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

#### DeepSeek 配置示例
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
      "llm": "main_llm",
      "memory": "chat_memory"
    }
  }
}
```

#### 智谱AI 完整配置示例 🆕
```json
{
  "llms": {
    "zhipu_llm": {
      "type": "zhipu",
      "model": "glm-4",
      "api_key": "${ZHIPU_API_KEY}",
      "temperature": 0.9,
      "max_tokens": 1024
    }
  },
  "memories": {
    "chat_memory": {
      "type": "conversation_buffer",
      "max_messages": 20
    }
  },
  "chains": {
    "zhipu_chain": {
      "type": "conversation",
      "llm": "zhipu_llm", 
      "memory": "chat_memory"
    }
  }
}
```

#### 硅基流动 完整配置示例 🆕
```json
{
  "llms": {
    "siliconflow_llm": {
      "type": "siliconflow",
      "model": "Qwen/Qwen2-7B-Instruct",
      "api_key": "${SILICONFLOW_API_KEY}",
      "temperature": 0.7,
      "max_tokens": 2048
    }
  },
  "memories": {
    "chat_memory": {
      "type": "conversation_buffer",
      "max_messages": 15
    }
  },
  "chains": {
    "sf_chain": {
      "type": "conversation",
      "llm": "siliconflow_llm",
      "memory": "chat_memory"
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

#### OpenAI 配置示例
```json
{
  "type": "openai",           // 必需：LLM 类型
  "model": "gpt-4",          // 必需：模型名称
  "api_key": "${OPENAI_API_KEY}",   // API 密钥（支持环境变量）
  "base_url": "https://...", // 可选：自定义 API 基础 URL
  "temperature": 0.7,        // 可选：温度参数
  "max_tokens": 2048,        // 可选：最大 Token 数
  "options": {               // 可选：其他选项
    "organization": "org-id"
  }
}
```

#### 智谱AI 配置示例 🆕
```json
{
  "type": "zhipu",
  "model": "glm-4",
  "api_key": "${ZHIPU_API_KEY}",
  "temperature": 0.9,
  "max_tokens": 1024
}
```

#### 硅基流动 配置示例 🆕
```json
{
  "type": "siliconflow", 
  "model": "Qwen/Qwen2-7B-Instruct",
  "api_key": "${SILICONFLOW_API_KEY}",
  "temperature": 0.7,
  "max_tokens": 2048
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
export ZHIPU_API_KEY="your-zhipu-key"               # 智谱AI 🆕
export SILICONFLOW_API_KEY="your-siliconflow-key"   # 硅基流动 🆕
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

### 配置生成器函数

#### 快速生成方法
- `QuickGenerateLLM(llmType, model, filename string) error`: 快速生成LLM配置
- `QuickGenerateChain(chainType, llmType, model, filename string) error`: 快速生成Chain配置
- `QuickGenerateAgent(agentType, llmType, model, filename string) error`: 快速生成Agent配置
- `QuickGenerateExecutor(agentType, llmType, model, filename string) error`: 快速生成Executor配置

#### 预设配置方法
- `GenerateDeepSeekChatConfig(filename string) error`: 生成DeepSeek聊天配置
- `GenerateKimiChatConfig(filename string) error`: 生成Kimi聊天配置
- `GenerateOpenAIChatConfig(filename string) error`: 生成OpenAI聊天配置
- `GenerateQwenChatConfig(filename string) error`: 生成通义千问聊天配置
- `GenerateReactAgentConfig(llmType, model, filename string) error`: 生成ReAct智能体配置
- `GenerateExecutorWithDeepSeek(filename string) error`: 生成DeepSeek执行器配置

#### 智谱AI配置方法 🆕
- `GenerateZhipuChatConfig(filename string) error`: 生成智谱AI聊天配置
- `GenerateExecutorWithZhipu(filename string) error`: 生成智谱AI执行器配置

#### 硅基流动配置方法 🆕
- `GenerateSiliconFlowChatConfig(filename string) error`: 生成硅基流动聊天配置
- `GenerateExecutorWithSiliconFlow(filename string) error`: 生成硅基流动执行器配置

#### 自定义配置方法
- `GenerateLLMConfig(template LLMTemplate, filename string) error`: 自定义LLM配置
- `GenerateChainConfig(template ChainTemplate, filename string) error`: 自定义Chain配置
- `GenerateAgentConfig(template AgentTemplate, filename string) error`: 自定义Agent配置
- `GenerateExecutorConfig(template ExecutorTemplate, filename string) error`: 自定义Executor配置

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个包。请确保：

1. 添加适当的测试用例
2. 更新文档
3. 遵循现有的代码风格

## 许可证

本项目采用与主项目相同的许可证。