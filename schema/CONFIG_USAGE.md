# 配置使用风格 (Config Usage Style)

这个模块提供了一种新的配置风格，允许以组件为入口点，直接嵌入所有依赖的配置，而不是使用引用（ref）方式。

## 特点

### 原始配置风格 vs 新配置风格

**原始配置风格（引用方式）：**
```json
{
  "llms": {
    "main_llm": {
      "type": "deepseek",
      "model": "deepseek-chat",
      "api_key": "${DEEPSEEK_API_KEY}"
    }
  },
  "agents": {
    "main_agent": {
      "type": "zero_shot_react",
      "llm_ref": "main_llm"
    }
  },
  "executors": {
    "main_executor": {
      "agent_ref": "main_agent"
    }
  }
}
```

**新配置风格（直接嵌入）：**
```json
{
  "agent": {
    "type": "zero_shot_react",
    "llm": {
      "type": "deepseek",
      "model": "deepseek-chat",
      "api_key": "${DEEPSEEK_API_KEY}"
    }
  }
}
```

## 支持的入口点

### 1. Executor 入口点 (`ExecutorUsageConfig`)

以 Executor 为顶层入口，包含完整的 Agent 配置和依赖：

```go
type ExecutorUsageConfig struct {
    Agent                   *AgentUsageConfig
    Memory                  *MemoryUsageConfig
    MaxIterations           *int
    ReturnIntermediateSteps *bool
    ErrorHandler            *ErrorHandlerConfig
    Options                 map[string]interface{}
}
```

### 2. Chain 入口点 (`ChainUsageConfig`)

以 Chain 为顶层入口，支持复杂的链式处理：

```go
type ChainUsageConfig struct {
    Type           string
    LLM            *LLMConfig
    Memory         *MemoryUsageConfig
    Prompt         *PromptConfig
    Chains         []*ChainUsageConfig  // 支持嵌套链
    // ... 其他字段
}
```

## 使用方法

### 1. 从 JSON 文件加载

```go
// 加载 Executor 配置
config, err := schema.LoadExecutorUsageConfigFromFile("executor_config.json")
if err != nil {
    log.Fatal(err)
}

// 加载 Chain 配置
chainConfig, err := schema.LoadChainUsageConfigFromFile("chain_config.json")
if err != nil {
    log.Fatal(err)
}
```

### 2. 从 JSON 字符串加载

```go
jsonStr := `{
  "agent": {
    "type": "zero_shot_react",
    "llm": {
      "type": "deepseek",
      "model": "deepseek-chat",
      "api_key": "${DEEPSEEK_API_KEY}"
    }
  }
}`

config, err := schema.LoadExecutorUsageConfigFromJSON(jsonStr)
```

### 3. 程序化创建

```go
config := &schema.ExecutorUsageConfig{
    Agent: &schema.AgentUsageConfig{
        Type: "zero_shot_react",
        LLM: &schema.LLMConfig{
            Type:    "deepseek",
            Model:   "deepseek-chat",
            APIKey:  "${DEEPSEEK_API_KEY}",
        },
    },
    MaxIterations: intPtr(10),
}
```

### 4. 配置验证

```go
if err := config.Validate(); err != nil {
    log.Fatalf("配置验证失败: %v", err)
}
```

### 5. 转换为原始配置格式

```go
// 转换为原始的 Config 格式，用于与现有工厂系统集成
originalConfig, err := config.ToConfig()
if err != nil {
    log.Fatal(err)
}

// 现在可以使用原始的工厂系统创建组件
factory := schema.NewFactory()
app, err := factory.CreateApplication(originalConfig)
```

## 配置层次结构

### Executor 配置层次：
```
ExecutorUsageConfig
├── AgentUsageConfig
│   ├── LLMConfig
│   └── MemoryUsageConfig
│       └── LLMConfig (可选，用于 summary 类型)
├── MemoryUsageConfig (可选，独立的 executor memory)
│   └── LLMConfig (可选)
└── ErrorHandlerConfig (可选)
```

### Chain 配置层次：
```
ChainUsageConfig
├── LLMConfig (可选)
├── MemoryUsageConfig (可选)
│   └── LLMConfig (可选)
├── PromptConfig (可选)
└── Chains[] (可选，用于 sequential 类型)
    └── ChainUsageConfig (递归嵌套)
```

## 示例文件

- `simple_executor.json` - 简单的 Executor 配置
- `executor_usage.json` - 完整的 Executor 配置示例
- `chain_usage.json` - Sequential Chain 配置示例
- `usage_config_example.go` - Go 代码使用示例

## 优势

1. **简化配置**: 不需要管理复杂的引用关系
2. **自包含**: 每个配置文件包含所有必需的依赖
3. **易于理解**: 配置结构直观反映组件层次关系
4. **减少错误**: 避免引用不存在的组件
5. **便于版本控制**: 配置变更集中在单一文件中

## 兼容性

新配置风格与原始配置系统完全兼容：

- 可以通过 `ToConfig()` 方法转换为原始格式
- 使用相同的验证规则和错误处理
- 支持所有现有的组件类型和选项

## 适用场景

新配置风格特别适合以下场景：

1. **快速原型开发**: 快速定义和测试 AI 应用
2. **单一用途应用**: 不需要复杂的组件复用
3. **配置模板**: 创建可复用的配置模板
4. **学习和演示**: 更容易理解组件关系