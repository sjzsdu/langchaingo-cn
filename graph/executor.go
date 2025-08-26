// Package graph - Execution engine implementation
// 包 graph - 执行引擎实现
package graph

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ================================
// Execution Engine 执行引擎
// ================================

// Runnable represents a compiled and executable graph.
// Runnable 表示已编译且可执行的图。
type Runnable struct {
	// graph is the underlying graph instance.
	graph *Graph

	// executionStats tracks execution statistics.
	executionStats *ExecutionStats

	// lock protects concurrent access.
	lock sync.RWMutex
}

// ExecutionStats tracks statistics about graph execution.
// ExecutionStats 跟踪图执行的统计信息。
type ExecutionStats struct {
	// TotalExecutions is the total number of executions.
	TotalExecutions int64 `json:"total_executions"`

	// SuccessfulExecutions is the number of successful executions.
	SuccessfulExecutions int64 `json:"successful_executions"`

	// FailedExecutions is the number of failed executions.
	FailedExecutions int64 `json:"failed_executions"`

	// AverageExecutionTime is the average execution time.
	AverageExecutionTime time.Duration `json:"average_execution_time"`

	// LastExecutionTime is the time of the last execution.
	LastExecutionTime time.Time `json:"last_execution_time"`

	// NodeExecutionCount tracks how many times each node has been executed.
	NodeExecutionCount map[string]int64 `json:"node_execution_count"`

	// NodeExecutionTime tracks total execution time for each node.
	NodeExecutionTime map[string]time.Duration `json:"node_execution_time"`

	// lock protects concurrent access to stats.
	lock sync.RWMutex
}

// ExecutionContext contains context information for a graph execution.
// ExecutionContext 包含图执行的上下文信息。
type ExecutionContext struct {
	// ExecutionID is a unique identifier for this execution.
	ExecutionID string

	// StartTime is when the execution started.
	StartTime time.Time

	// Timeout is the maximum execution time.
	Timeout time.Duration

	// MaxSteps is the maximum number of steps to execute.
	MaxSteps int

	// EnableTracing enables detailed execution tracing.
	EnableTracing bool

	// Context is the underlying Go context.
	Context context.Context

	// Cancel is the cancellation function.
	Cancel context.CancelFunc

	// StepCount tracks the number of steps executed.
	StepCount int

	// Trace contains execution trace information.
	Trace []TraceEntry
}

// TraceEntry represents a single trace entry.
// TraceEntry 表示单个追踪条目。
type TraceEntry struct {
	// Timestamp is when this trace entry was created.
	Timestamp time.Time `json:"timestamp"`

	// NodeID is the ID of the node being traced.
	NodeID string `json:"node_id"`

	// Event is the type of event (start, end, error, etc.).
	Event string `json:"event"`

	// Message is a human-readable message.
	Message string `json:"message"`

	// Data contains additional trace data.
	Data map[string]interface{} `json:"data,omitempty"`
}

// ================================
// Execution Options 执行选项
// ================================

// ExecutionOption represents an option for graph execution.
// ExecutionOption 表示图执行的选项。
type ExecutionOption func(*ExecutionContext)

// WithTimeout sets the execution timeout.
// WithTimeout 设置执行超时时间。
func WithTimeout(timeout time.Duration) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.Timeout = timeout
	}
}

// WithMaxSteps sets the maximum number of steps to execute.
// WithMaxSteps 设置最大执行步数。
func WithMaxSteps(maxSteps int) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.MaxSteps = maxSteps
	}
}

// WithTracing enables detailed execution tracing.
// WithTracing 启用详细的执行追踪。
func WithTracing(enabled bool) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.EnableTracing = enabled
	}
}

// WithExecutionID sets a custom execution ID.
// WithExecutionID 设置自定义执行ID。
func WithExecutionID(id string) ExecutionOption {
	return func(ctx *ExecutionContext) {
		ctx.ExecutionID = id
	}
}

// ================================
// Main Execution Methods 主要执行方法
// ================================

// Invoke executes the graph with the given messages.
// Invoke 使用给定的消息执行图。
func (r *Runnable) Invoke(ctx context.Context, state *State) (*State, error) {
	return r.InvokeWithOptions(ctx, state)
}

// InvokeWithOptions executes the graph with the given state and options.
// InvokeWithOptions 使用给定的状态和选项执行图。
func (r *Runnable) InvokeWithOptions(ctx context.Context, state *State, options ...ExecutionOption) (*State, error) {
	// Create execution context
	execCtx := &ExecutionContext{
		ExecutionID:   fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		StartTime:     time.Now(),
		Timeout:       r.graph.Config.Timeout,
		MaxSteps:      1000, // Default max steps
		EnableTracing: false,
		Context:       ctx,
		StepCount:     0,
		Trace:         make([]TraceEntry, 0),
	}

	// Apply options
	for _, option := range options {
		option(execCtx)
	}

	// Create context with timeout
	if execCtx.Timeout > 0 {
		execCtx.Context, execCtx.Cancel = context.WithTimeout(ctx, execCtx.Timeout)
	} else {
		execCtx.Context, execCtx.Cancel = context.WithCancel(ctx)
	}
	defer execCtx.Cancel()

	// Record execution start
	r.recordExecutionStart(execCtx)

	// Execute the graph
	result, err := r.executeGraph(execCtx, state)

	// Record execution end
	r.recordExecutionEnd(execCtx, err)

	return result, err
}

// executeGraph executes the graph starting from the entry point.
// executeGraph 从入口点开始执行图。
func (r *Runnable) executeGraph(execCtx *ExecutionContext, state *State) (*State, error) {
	currentNodeID := r.graph.entryPoint
	currentState := state.Clone()

	for {
		// Check context cancellation
		select {
		case <-execCtx.Context.Done():
			return nil, execCtx.Context.Err()
		default:
		}

		// Check max steps
		if execCtx.MaxSteps > 0 && execCtx.StepCount >= execCtx.MaxSteps {
			return nil, fmt.Errorf("maximum execution steps (%d) exceeded", execCtx.MaxSteps)
		}

		// Check if we've reached the end
		if currentNodeID == "END" || currentNodeID == "" {
			break
		}

		// Get the current node
		node, exists := r.graph.GetNode(currentNodeID)
		if !exists {
			return nil, fmt.Errorf("node %s not found", currentNodeID)
		}

		// Trace node execution start
		if execCtx.EnableTracing {
			r.addTraceEntry(execCtx, currentNodeID, "node_start", "Starting node execution", nil)
		}

		// Execute the node
		nodeStartTime := time.Now()
		newState, err := r.executeNode(execCtx, node, currentState)
		nodeExecutionTime := time.Since(nodeStartTime)

		// Update node execution stats
		r.updateNodeStats(node.ID, nodeExecutionTime, err == nil)

		if err != nil {
			// Trace error
			if execCtx.EnableTracing {
				r.addTraceEntry(execCtx, currentNodeID, "node_error", "Node execution failed", map[string]interface{}{
					"error": err.Error(),
				})
			}

			// Handle error based on node failure mode
			switch node.Config.FailureMode {
			case FailureModeStop:
				return nil, fmt.Errorf("node %s failed: %w", currentNodeID, err)
			case FailureModeContinue, FailureModeSkip:
				// Continue with current state
				newState = currentState
			default:
				return nil, fmt.Errorf("node %s failed: %w", currentNodeID, err)
			}
		}

		// Trace node execution end
		if execCtx.EnableTracing {
			r.addTraceEntry(execCtx, currentNodeID, "node_end", "Node execution completed", map[string]interface{}{
				"duration_ms": nodeExecutionTime.Milliseconds(),
				"success":     err == nil,
			})
		}

		// Update state
		currentState = newState
		execCtx.StepCount++

		// Determine next node
		nextNodeID, err := r.graph.router.GetNextNode(execCtx.Context, currentNodeID, currentState)
		if err != nil {
			return nil, fmt.Errorf("failed to determine next node from %s: %w", currentNodeID, err)
		}

		// Trace routing decision
		if execCtx.EnableTracing {
			r.addTraceEntry(execCtx, currentNodeID, "routing", "Routing to next node", map[string]interface{}{
				"next_node": nextNodeID,
			})
		}

		currentNodeID = nextNodeID
	}

	return currentState, nil
}

// executeNode executes a single node with middleware support.
// executeNode 执行单个节点，支持中间件。
func (r *Runnable) executeNode(execCtx *ExecutionContext, node *Node, state *State) (*State, error) {
	// Create final execution function
	finalFunc := func(ctx context.Context, state *State) (*State, error) {
		return node.Execute(ctx, state)
	}

	// Apply graph-level middleware
	for i := len(r.graph.middleware) - 1; i >= 0; i-- {
		middleware := r.graph.middleware[i]
		prevFunc := finalFunc
		finalFunc = func(ctx context.Context, state *State) (*State, error) {
			return middleware.Process(ctx, prevFunc, state)
		}
	}

	return finalFunc(execCtx.Context, state)
}

// ================================
// Parallel Execution 并行执行
// ================================

// InvokeParallel executes multiple instances of the graph in parallel.
// InvokeParallel 并行执行图的多个实例。
func (r *Runnable) InvokeParallel(ctx context.Context, states []*State, options ...ExecutionOption) ([]*Result, error) {
	if len(states) == 0 {
		return nil, fmt.Errorf("no states provided for parallel execution")
	}

	maxConcurrency := r.graph.Config.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = len(states)
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)
	results := make([]*Result, len(states))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, state := range states {
		wg.Add(1)
		go func(index int, inputState *State) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			startTime := time.Now()
			finalState, err := r.InvokeWithOptions(ctx, inputState, options...)
			duration := time.Since(startTime)

			result := &Result{
				State:    finalState,
				Success:  err == nil,
				Error:    err,
				Duration: duration,
			}

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, state)
	}

	wg.Wait()
	return results, nil
}

// ================================
// Streaming Execution 流式执行
// ================================

// Stream executes the graph and streams intermediate results.
// Stream 执行图并流式传输中间结果。
func (r *Runnable) Stream(ctx context.Context, state *State, options ...ExecutionOption) (<-chan *StreamResult, error) {
	resultChan := make(chan *StreamResult, 100)

	go func() {
		defer close(resultChan)

		// Create execution context with tracing enabled
		opts := append(options, WithTracing(true))
		finalState, err := r.InvokeWithOptions(ctx, state, opts...)

		if err != nil {
			resultChan <- &StreamResult{
				Type:  StreamResultTypeError,
				Error: err,
			}
			return
		}

		// Send final result
		resultChan <- &StreamResult{
			Type:  StreamResultTypeFinal,
			State: finalState,
		}
	}()

	return resultChan, nil
}

// StreamResult represents a streaming execution result.
// StreamResult 表示流式执行结果。
type StreamResult struct {
	// Type indicates the type of this result.
	Type StreamResultType

	// State contains the current state (for intermediate and final results).
	State *State

	// NodeID is the ID of the node that produced this result.
	NodeID string

	// Error contains any error that occurred.
	Error error

	// Metadata contains additional result metadata.
	Metadata map[string]interface{}
}

// StreamResultType represents the type of a streaming result.
// StreamResultType 表示流式结果的类型。
type StreamResultType string

const (
	// StreamResultTypeIntermediate represents an intermediate result.
	StreamResultTypeIntermediate StreamResultType = "intermediate"
	// StreamResultTypeFinal represents the final result.
	StreamResultTypeFinal StreamResultType = "final"
	// StreamResultTypeError represents an error result.
	StreamResultTypeError StreamResultType = "error"
)

// ================================
// Statistics and Monitoring 统计和监控
// ================================

// recordExecutionStart records the start of an execution.
// recordExecutionStart 记录执行的开始。
func (r *Runnable) recordExecutionStart(execCtx *ExecutionContext) {
	if r.executionStats == nil {
		r.executionStats = &ExecutionStats{
			NodeExecutionCount: make(map[string]int64),
			NodeExecutionTime:  make(map[string]time.Duration),
		}
	}

	r.executionStats.lock.Lock()
	defer r.executionStats.lock.Unlock()

	r.executionStats.TotalExecutions++
	r.executionStats.LastExecutionTime = execCtx.StartTime
}

// recordExecutionEnd records the end of an execution.
// recordExecutionEnd 记录执行的结束。
func (r *Runnable) recordExecutionEnd(execCtx *ExecutionContext, err error) {
	r.executionStats.lock.Lock()
	defer r.executionStats.lock.Unlock()

	duration := time.Since(execCtx.StartTime)

	if err == nil {
		r.executionStats.SuccessfulExecutions++
	} else {
		r.executionStats.FailedExecutions++
	}

	// Update average execution time
	totalExecs := r.executionStats.TotalExecutions
	if totalExecs > 0 {
		currentAvg := r.executionStats.AverageExecutionTime
		r.executionStats.AverageExecutionTime = (currentAvg*time.Duration(totalExecs-1) + duration) / time.Duration(totalExecs)
	}
}

// updateNodeStats updates statistics for a specific node.
// updateNodeStats 更新特定节点的统计信息。
func (r *Runnable) updateNodeStats(nodeID string, duration time.Duration, success bool) {
	if r.executionStats == nil {
		return
	}

	r.executionStats.lock.Lock()
	defer r.executionStats.lock.Unlock()

	r.executionStats.NodeExecutionCount[nodeID]++
	r.executionStats.NodeExecutionTime[nodeID] += duration
}

// addTraceEntry adds a trace entry to the execution context.
// addTraceEntry 向执行上下文添加追踪条目。
func (r *Runnable) addTraceEntry(execCtx *ExecutionContext, nodeID, event, message string, data map[string]interface{}) {
	entry := TraceEntry{
		Timestamp: time.Now(),
		NodeID:    nodeID,
		Event:     event,
		Message:   message,
		Data:      data,
	}
	execCtx.Trace = append(execCtx.Trace, entry)
}

// GetExecutionStats returns the current execution statistics.
// GetExecutionStats 返回当前的执行统计信息。
func (r *Runnable) GetExecutionStats() *ExecutionStats {
	if r.executionStats == nil {
		return nil
	}

	r.executionStats.lock.RLock()
	defer r.executionStats.lock.RUnlock()

	// Return a copy to avoid concurrent access issues
	stats := &ExecutionStats{
		TotalExecutions:      r.executionStats.TotalExecutions,
		SuccessfulExecutions: r.executionStats.SuccessfulExecutions,
		FailedExecutions:     r.executionStats.FailedExecutions,
		AverageExecutionTime: r.executionStats.AverageExecutionTime,
		LastExecutionTime:    r.executionStats.LastExecutionTime,
		NodeExecutionCount:   make(map[string]int64),
		NodeExecutionTime:    make(map[string]time.Duration),
	}

	for k, v := range r.executionStats.NodeExecutionCount {
		stats.NodeExecutionCount[k] = v
	}

	for k, v := range r.executionStats.NodeExecutionTime {
		stats.NodeExecutionTime[k] = v
	}

	return stats
}

// ResetStats resets all execution statistics.
// ResetStats 重置所有执行统计信息。
func (r *Runnable) ResetStats() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.executionStats = &ExecutionStats{
		NodeExecutionCount: make(map[string]int64),
		NodeExecutionTime:  make(map[string]time.Duration),
	}
}

// ================================
// Utility Methods 工具方法
// ================================

// GetGraph returns the underlying graph.
// GetGraph 返回底层图。
func (r *Runnable) GetGraph() *Graph {
	return r.graph
}

// GetExecutionTrace returns the execution trace from the last execution.
// GetExecutionTrace 返回最后一次执行的执行追踪。
func (r *Runnable) GetExecutionTrace() []TraceEntry {
	// This would require storing the last execution context
	// For now, return an empty slice
	return make([]TraceEntry, 0)
}