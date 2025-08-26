# Graph Framework - 图框架

一个强大、灵活的Go语言图处理框架，专为构建复杂的工作流和状态机而设计。

## 特性 Features

### 🏗️ 核心功能 Core Features
- **节点系统** - 支持多种节点类型（函数、条件、并行、循环、子图等）
- **边路由** - 灵活的条件路由和优先级路由
- **状态管理** - 完整的状态生命周期管理和持久化
- **执行引擎** - 支持串行、并行和流式执行
- **中间件** - 丰富的中间件生态系统

### 🚀 高级特性 Advanced Features
- **链式构建** - 流畅的API设计，支持链式调用
- **条件路由** - 基于状态的智能路由决策
- **并发执行** - 支持并行节点执行和并发控制
- **错误处理** - 多种错误处理策略（停止、继续、重试、跳过）
- **监控和追踪** - 内置执行统计和详细追踪
- **状态持久化** - 多种存储后端（内存、文件、Redis等）

### 🔧 中间件生态 Middleware Ecosystem
- **日志中间件** - 详细的执行日志记录
- **指标中间件** - 性能指标收集
- **超时中间件** - 执行超时控制
- **重试中间件** - 智能重试机制
- **断路器中间件** - 故障保护
- **限流中间件** - 请求限流
- **验证中间件** - 输入输出验证

## 快速开始 Quick Start

### 基本使用 Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/sjzsdu/langchaingo-cn/graph"
    "github.com/tmc/langchaingo/llms"
)

func main() {
    // 创建节点
    processNode := graph.NewNode("process").
        WithName("处理节点").
        WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
            // 处理逻辑
            state.AddMessage(llms.MessageContent{
                Role: llms.ChatMessageTypeAI,
                Parts: []llms.ContentPart{llms.TextPart("处理完成")},
            })
            return state, nil
        }).
        Build()

    endNode := graph.NewNode("END").
        WithType(graph.NodeTypeEnd).
        Build()

    // 创建图
    g := graph.NewGraph("simple_graph").
        WithName("简单图").
        AddNodes(processNode, endNode).
        Connect("process", "END").
        SetEntryPoint("process").
        Build()

    // 编译并执行
    runnable, err := g.Compile()
    if err != nil {
        panic(err)
    }

    state := graph.NewState("example")
    result, err := runnable.Invoke(context.Background(), state)
    if err != nil {
        panic(err)
    }

    fmt.Printf("执行完成，消息数量: %d\n", len(result.Messages))
}
```

### 条件路由 Conditional Routing

```go
// 创建条件节点
classifyNode := graph.NewNode("classify").
    WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
        // 分类逻辑
        if len(state.Messages) > 0 {
            text := getTextFromMessage(state.Messages[len(state.Messages)-1])
            if strings.Contains(text, "问题") {
                state.SetVariable("type", "question")
            } else {
                state.SetVariable("type", "general")
            }
        }
        return state, nil
    }).
    Build()

// 条件边
questionEdge := graph.NewEdge("to_question", "classify", "handle_question").
    WithCondition(func(ctx context.Context, state *graph.State) (bool, error) {
        t, exists := state.GetVariable("type")
        return exists && t == "question", nil
    }).
    Build()

generalEdge := graph.NewEdge("to_general", "classify", "handle_general").
    AsDefault(). // 默认路径
    Build()
```

### 中间件使用 Middleware Usage

```go
// 创建中间件
loggingMiddleware := graph.NewLoggingMiddleware(graph.LogLevelInfo)
metricsMiddleware := graph.NewMetricsMiddleware()
timeoutMiddleware := graph.NewTimeoutMiddleware(30 * time.Second)

// 应用到节点
node := graph.NewNode("node_with_middleware").
    WithFunction(nodeFunction).
    WithMiddleware(loggingMiddleware, metricsMiddleware).
    Build()

// 应用到图
g := graph.NewGraph("graph_with_middleware").
    WithMiddleware(timeoutMiddleware).
    AddNode(node).
    Build()
```

### 状态管理 State Management

```go
// 内存状态管理器
memoryManager := graph.NewMemoryStateManager(1000)

// 文件状态管理器
fileManager, err := graph.NewFileStateManager("./states")
if err != nil {
    panic(err)
}

// 复合状态管理器
compositeManager := graph.NewCompositeStateManager(
    memoryManager,     // 主要
    fileManager,       // 备份
    true,              // 写通
    graph.ReadPrimaryFirst,
)

// 应用到图
g := graph.NewGraph("stateful_graph").
    WithStateManager(compositeManager).
    Build()
```

## 架构设计 Architecture

### 核心组件 Core Components

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    Node     │    │    Edge     │    │   Graph     │
│   节点      │    │    边       │    │    图       │
├─────────────┤    ├─────────────┤    ├─────────────┤
│ • Function  │    │ • Condition │    │ • Nodes     │
│ • Condition │    │ • Priority  │    │ • Edges     │
│ • Parallel  │    │ • Weight    │    │ • Router    │
│ • Loop      │    │ • Metadata  │    │ • Config    │
│ • SubGraph  │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
                  ┌─────────────┐
                  │  Runnable   │
                  │  可执行实例  │
                  ├─────────────┤
                  │ • Executor  │
                  │ • Stats     │
                  │ • Tracing   │
                  └─────────────┘
```

### 执行流程 Execution Flow

```
开始 → 验证图 → 编译图 → 创建执行上下文 → 执行节点 → 路由决策 → 下个节点 → 结束
 ↓         ↓         ↓            ↓            ↓         ↓         ↓         ↓
状态初始化 → 依赖检查 → 优化处理 → 中间件加载 → 函数执行 → 条件评估 → 状态更新 → 结果返回
```

## 节点类型 Node Types

### 函数节点 Function Node
```go
funcNode := graph.NewNode("func").
    WithType(graph.NodeTypeFunction).
    WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
        // 自定义处理逻辑
        return state, nil
    }).
    Build()
```

### 条件节点 Condition Node
```go
condNode := graph.NewNode("condition").
    WithType(graph.NodeTypeCondition).
    WithCondition(func(ctx context.Context, state *graph.State) (string, error) {
        // 返回下一个节点ID
        return "next_node", nil
    }).
    Build()
```

### 并行节点 Parallel Node
```go
parallelNode := graph.NewNode("parallel").
    WithType(graph.NodeTypeParallel).
    // 并行执行配置
    Build()
```

## 边类型 Edge Types

### 普通边 Normal Edge
```go
edge := graph.NewEdge("normal", "from", "to").Build()
```

### 条件边 Conditional Edge
```go
condEdge := graph.NewEdge("conditional", "from", "to").
    WithCondition(func(ctx context.Context, state *graph.State) (bool, error) {
        return true, nil
    }).
    Build()
```

### 优先级边 Priority Edge
```go
priorityEdge := graph.NewEdge("priority", "from", "to").
    WithPriority(10).
    Build()
```

### 默认边 Default Edge
```go
defaultEdge := graph.NewEdge("default", "from", "to").
    AsDefault().
    Build()
```

## 中间件详解 Middleware Details

### 日志中间件 Logging Middleware
```go
logging := graph.NewLoggingMiddleware(graph.LogLevelInfo)
logging.IncludeState = true // 包含状态信息
```

### 指标中间件 Metrics Middleware
```go
metrics := graph.NewMetricsMiddleware()

// 获取指标
allMetrics := metrics.GetMetrics()
nodeMetrics, exists := metrics.GetNodeMetrics("node_id")
```

### 超时中间件 Timeout Middleware
```go
timeout := graph.NewTimeoutMiddleware(30 * time.Second)
timeout.SetNodeTimeout("slow_node", 60 * time.Second)
```

### 重试中间件 Retry Middleware
```go
retry := graph.NewRetryMiddleware(3, 1*time.Second)
retry.BackoffMultiplier = 2.0
retry.ShouldRetry = func(err error) bool {
    return !isNetworkError(err)
}
```

### 断路器中间件 Circuit Breaker Middleware
```go
cb := graph.NewCircuitBreakerMiddleware(5, 30*time.Second)
state := cb.GetState() // 获取断路器状态
```

## 状态管理 State Management

### 内存状态管理器 Memory State Manager
```go
memory := graph.NewMemoryStateManager(1000) // 最多1000个状态
stats := memory.GetStats()
memory.Clear() // 清空所有状态
```

### 文件状态管理器 File State Manager
```go
file, err := graph.NewFileStateManager("./state_files")
stateIDs, err := file.ListStates()
err = file.Cleanup(24 * time.Hour) // 清理24小时前的状态
```

### 检查点管理器 Checkpoint Manager
```go
checkpoints := graph.NewCheckpointManager(
    stateManager, 
    5*time.Minute,  // 检查点间隔
    10,             // 最多保留10个检查点
)

// 创建检查点
err = checkpoints.CreateCheckpoint(ctx, state)

// 恢复检查点
restoredState, err := checkpoints.RestoreFromCheckpoint(ctx, "state_id", 0)
```

## 执行选项 Execution Options

```go
result, err := runnable.InvokeWithOptions(ctx, state,
    graph.WithTimeout(30*time.Second),     // 设置超时
    graph.WithMaxSteps(100),               // 最大步数
    graph.WithTracing(true),               // 启用追踪
    graph.WithExecutionID("custom_id"),    // 自定义执行ID
)
```

## 并行执行 Parallel Execution

```go
states := []*graph.State{state1, state2, state3}
results, err := runnable.InvokeParallel(ctx, states,
    graph.WithTimeout(60*time.Second),
)

for i, result := range results {
    if result.Success {
        fmt.Printf("State %d executed successfully\n", i)
    } else {
        fmt.Printf("State %d failed: %v\n", i, result.Error)
    }
}
```

## 流式执行 Streaming Execution

```go
stream, err := runnable.Stream(ctx, state)
if err != nil {
    panic(err)
}

for result := range stream {
    switch result.Type {
    case graph.StreamResultTypeIntermediate:
        fmt.Printf("Intermediate result from %s\n", result.NodeID)
    case graph.StreamResultTypeFinal:
        fmt.Println("Final result received")
    case graph.StreamResultTypeError:
        fmt.Printf("Error: %v\n", result.Error)
    }
}
```

## 图验证 Graph Validation

```go
validation := graph.Validate()
if !validation.Valid {
    fmt.Println("图验证失败:")
    for _, err := range validation.Errors {
        fmt.Printf("  错误: %s\n", err.Message)
    }
}

for _, warning := range validation.Warnings {
    fmt.Printf("  警告: %s\n", warning.Message)
}
```

## 性能监控 Performance Monitoring

```go
// 执行统计
stats := runnable.GetExecutionStats()
fmt.Printf("总执行次数: %d\n", stats.TotalExecutions)
fmt.Printf("成功次数: %d\n", stats.SuccessfulExecutions)
fmt.Printf("平均执行时间: %v\n", stats.AverageExecutionTime)

// 节点统计
for nodeID, count := range stats.NodeExecutionCount {
    duration := stats.NodeExecutionTime[nodeID]
    avgDuration := duration / time.Duration(count)
    fmt.Printf("节点 %s: %d次执行, 平均耗时 %v\n", nodeID, count, avgDuration)
}
```

## 最佳实践 Best Practices

### 1. 图设计原则
- **单一职责**: 每个节点专注于单一功能
- **最小依赖**: 减少节点间的强耦合
- **明确边界**: 清晰定义输入输出契约
- **错误隔离**: 合理设置错误处理策略

### 2. 性能优化
- **并行执行**: 利用并行节点提高吞吐量
- **缓存状态**: 合理使用状态缓存
- **资源池**: 复用昂贵的资源（如LLM连接）
- **批处理**: 批量处理相似任务

### 3. 监控和调试
- **启用追踪**: 在开发环境启用详细追踪
- **指标收集**: 监控关键性能指标
- **日志记录**: 记录重要的执行信息
- **检查点**: 对长时间运行的任务设置检查点

### 4. 错误处理
- **优雅降级**: 设计合理的降级策略
- **重试机制**: 对临时错误进行重试
- **断路器**: 保护下游服务
- **告警机制**: 及时发现和响应问题

## 与原实现的对比 Comparison with Original

| 特性 | 原实现 | 新框架 |
|------|--------|--------|
| 节点类型 | 单一函数节点 | 7种节点类型 |
| 路由机制 | 简单线性路由 | 条件路由+优先级 |
| 错误处理 | 基本错误传播 | 4种错误策略 |
| 中间件 | 无 | 7种内置中间件 |
| 状态管理 | 无持久化 | 多种存储后端 |
| 并发支持 | 无 | 完整并发控制 |
| 监控追踪 | 无 | 详细统计和追踪 |
| API设计 | 过程式 | 流畅的链式API |

## 示例项目 Example Projects

查看 `examples/` 目录下的完整示例：
- `basic_chat.go` - 基本聊天处理流程
- `advanced_workflow.go` - 高级工作流演示

## 测试 Testing

```bash
# 运行所有测试
go test ./graph/...

# 运行基准测试
go test -bench=. ./graph/...

# 运行特定测试
go test -run TestGraphExecution ./graph/...
```

## 贡献 Contributing

欢迎提交Pull Request和Issue！请确保：
1. 代码符合Go代码规范
2. 添加适当的测试用例
3. 更新相关文档

## 许可证 License

MIT License - 详见LICENSE文件

## 更新日志 Changelog

### v2.0.0
- 全新的图框架设计
- 支持多种节点类型
- 条件路由系统
- 丰富的中间件生态
- 完整的状态管理
- 并发执行支持
- 性能监控和追踪