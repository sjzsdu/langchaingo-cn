// Package graph provides a flexible and powerful graph-based workflow execution framework.
// 包 graph 提供了一个灵活且强大的基于图的工作流执行框架。
package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// ================================
// Core Interfaces 核心接口
// ================================

// Invokable represents any component that can be invoked with a context and state.
// Invokable 表示任何可以通过上下文和状态调用的组件。
type Invokable interface {
	Invoke(ctx context.Context, state *State) (*State, error)
}

// NodeProcessor defines the interface for processing nodes.
// NodeProcessor 定义了处理节点的接口。
type NodeProcessor interface {
	Process(ctx context.Context, node *Node, state *State) (*State, error)
}

// Middleware defines the interface for middleware components.
// Middleware 定义了中间件组件的接口。
type Middleware interface {
	Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error)
}

// StateManager defines the interface for managing graph execution state.
// StateManager 定义了管理图执行状态的接口。
type StateManager interface {
	Save(ctx context.Context, state *State) error
	Load(ctx context.Context, id string) (*State, error)
	Delete(ctx context.Context, id string) error
}

// ================================
// Core Types 核心类型
// ================================

// NodeType represents the type of a node.
// NodeType 表示节点的类型。
type NodeType string

const (
	// NodeTypeFunction represents a function node.
	NodeTypeFunction NodeType = "function"
	// NodeTypeCondition represents a condition node.
	NodeTypeCondition NodeType = "condition"
	// NodeTypeParallel represents a parallel execution node.
	NodeTypeParallel NodeType = "parallel"
	// NodeTypeLoop represents a loop node.
	NodeTypeLoop NodeType = "loop"
	// NodeTypeSubGraph represents a sub-graph node.
	NodeTypeSubGraph NodeType = "subgraph"
	// NodeTypeStart represents the start node.
	NodeTypeStart NodeType = "start"
	// NodeTypeEnd represents the end node.
	NodeTypeEnd NodeType = "end"
)

// ExecutionMode represents how a node should be executed.
// ExecutionMode 表示节点应该如何执行。
type ExecutionMode string

const (
	// ExecutionModeSequential executes nodes sequentially.
	ExecutionModeSequential ExecutionMode = "sequential"
	// ExecutionModeParallel executes nodes in parallel.
	ExecutionModeParallel ExecutionMode = "parallel"
	// ExecutionModeConcurrent executes nodes concurrently with synchronization.
	ExecutionModeConcurrent ExecutionMode = "concurrent"
)

// State represents the execution state of the graph.
// State 表示图的执行状态。
type State struct {
	// ID is the unique identifier for this state instance.
	ID string `json:"id"`

	// Messages contains the current message content.
	Messages []llms.MessageContent `json:"messages"`

	// Variables contains custom variables that can be used across nodes.
	Variables map[string]interface{} `json:"variables"`

	// Metadata contains execution metadata.
	Metadata map[string]interface{} `json:"metadata"`

	// History tracks the execution history.
	History []ExecutionStep `json:"history"`

	// CurrentNode indicates the currently executing node.
	CurrentNode string `json:"current_node"`

	// CreatedAt indicates when this state was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt indicates when this state was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// ExecutionStep represents a single step in the execution history.
// ExecutionStep 表示执行历史中的单个步骤。
type ExecutionStep struct {
	// NodeID is the ID of the executed node.
	NodeID string `json:"node_id"`

	// StartTime is when the step started.
	StartTime time.Time `json:"start_time"`

	// EndTime is when the step completed.
	EndTime time.Time `json:"end_time"`

	// Duration is how long the step took.
	Duration time.Duration `json:"duration"`

	// Success indicates if the step completed successfully.
	Success bool `json:"success"`

	// Error contains any error that occurred.
	Error string `json:"error,omitempty"`

	// Input is the input state for this step.
	Input map[string]interface{} `json:"input,omitempty"`

	// Output is the output state for this step.
	Output map[string]interface{} `json:"output,omitempty"`
}

// NodeFunction represents a function that can be executed by a node.
// NodeFunction 表示可以被节点执行的函数。
type NodeFunction func(ctx context.Context, state *State) (*State, error)

// ConditionFunction represents a condition evaluation function.
// ConditionFunction 表示条件评估函数。
type ConditionFunction func(ctx context.Context, state *State) (string, error)

// NodeConfig contains configuration for a node.
// NodeConfig 包含节点的配置。
type NodeConfig struct {
	// Timeout specifies the maximum execution time for this node.
	Timeout time.Duration `json:"timeout,omitempty"`

	// Retries specifies the number of retry attempts on failure.
	Retries int `json:"retries,omitempty"`

	// RetryDelay specifies the delay between retry attempts.
	RetryDelay time.Duration `json:"retry_delay,omitempty"`

	// FailureMode specifies how to handle failures.
	FailureMode FailureMode `json:"failure_mode,omitempty"`

	// Middleware contains middleware to apply to this node.
	Middleware []string `json:"middleware,omitempty"`

	// Metadata contains custom metadata for this node.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FailureMode represents how to handle node failures.
// FailureMode 表示如何处理节点失败。
type FailureMode string

const (
	// FailureModeStop stops execution on failure.
	FailureModeStop FailureMode = "stop"
	// FailureModeContinue continues execution on failure.
	FailureModeContinue FailureMode = "continue"
	// FailureModeRetry retries the node on failure.
	FailureModeRetry FailureMode = "retry"
	// FailureModeSkip skips the node on failure and continues.
	FailureModeSkip FailureMode = "skip"
)

// GraphConfig contains configuration for the entire graph.
// GraphConfig 包含整个图的配置。
type GraphConfig struct {
	// Name is the name of the graph.
	Name string `json:"name"`

	// Description is a description of what this graph does.
	Description string `json:"description"`

	// Version is the version of this graph.
	Version string `json:"version"`

	// ExecutionMode specifies how nodes should be executed by default.
	ExecutionMode ExecutionMode `json:"execution_mode"`

	// MaxConcurrency specifies the maximum number of concurrent executions.
	MaxConcurrency int `json:"max_concurrency"`

	// Timeout specifies the maximum execution time for the entire graph.
	Timeout time.Duration `json:"timeout"`

	// EnableStateTracking enables state tracking and history.
	EnableStateTracking bool `json:"enable_state_tracking"`

	// StateManager specifies which state manager to use.
	StateManager string `json:"state_manager,omitempty"`

	// Metadata contains custom metadata for this graph.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ================================
// Helper Types 辅助类型
// ================================

// Result represents the result of a graph execution.
// Result 表示图执行的结果。
type Result struct {
	// State is the final execution state.
	State *State

	// Success indicates if the execution was successful.
	Success bool

	// Error contains any error that occurred.
	Error error

	// Duration is how long the execution took.
	Duration time.Duration

	// NodesExecuted is the number of nodes that were executed.
	NodesExecuted int
}

// ValidationResult represents the result of graph validation.
// ValidationResult 表示图验证的结果。
type ValidationResult struct {
	// Valid indicates if the graph is valid.
	Valid bool

	// Errors contains any validation errors.
	Errors []ValidationError

	// Warnings contains any validation warnings.
	Warnings []ValidationWarning
}

// ValidationError represents a validation error.
// ValidationError 表示验证错误。
type ValidationError struct {
	// Code is the error code.
	Code string

	// Message is the error message.
	Message string

	// NodeID is the ID of the node that caused the error (if applicable).
	NodeID string

	// Details contains additional error details.
	Details map[string]interface{}
}

// ValidationWarning represents a validation warning.
// ValidationWarning 表示验证警告。
type ValidationWarning struct {
	// Code is the warning code.
	Code string

	// Message is the warning message.
	Message string

	// NodeID is the ID of the node that caused the warning (if applicable).
	NodeID string

	// Details contains additional warning details.
	Details map[string]interface{}
}

// ================================
// Utility Functions 工具函数
// ================================

// NewState creates a new state instance.
// NewState 创建一个新的状态实例。
func NewState(id string) *State {
	now := time.Now()
	return &State{
		ID:          id,
		Messages:    make([]llms.MessageContent, 0),
		Variables:   make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
		History:     make([]ExecutionStep, 0),
		CurrentNode: "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Clone creates a deep copy of the state.
// Clone 创建状态的深拷贝。
func (s *State) Clone() *State {
	clone := &State{
		ID:          s.ID,
		Messages:    make([]llms.MessageContent, len(s.Messages)),
		Variables:   make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
		History:     make([]ExecutionStep, len(s.History)),
		CurrentNode: s.CurrentNode,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	copy(clone.Messages, s.Messages)
	copy(clone.History, s.History)

	for k, v := range s.Variables {
		clone.Variables[k] = v
	}

	for k, v := range s.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

// AddMessage adds a message to the state.
// AddMessage 向状态添加消息。
func (s *State) AddMessage(msg llms.MessageContent) {
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()
}

// SetVariable sets a variable in the state.
// SetVariable 在状态中设置变量。
func (s *State) SetVariable(key string, value interface{}) {
	s.Variables[key] = value
	s.UpdatedAt = time.Now()
}

// GetVariable gets a variable from the state.
// GetVariable 从状态获取变量。
func (s *State) GetVariable(key string) (interface{}, bool) {
	value, exists := s.Variables[key]
	return value, exists
}

// SetMetadata sets metadata in the state.
// SetMetadata 在状态中设置元数据。
func (s *State) SetMetadata(key string, value interface{}) {
	s.Metadata[key] = value
	s.UpdatedAt = time.Now()
}

// GetMetadata gets metadata from the state.
// GetMetadata 从状态获取元数据。
func (s *State) GetMetadata(key string) (interface{}, bool) {
	value, exists := s.Metadata[key]
	return value, exists
}

// AddExecutionStep adds an execution step to the history.
// AddExecutionStep 向历史记录添加执行步骤。
func (s *State) AddExecutionStep(step ExecutionStep) {
	s.History = append(s.History, step)
	s.UpdatedAt = time.Now()
}

// String returns a string representation of the node type.
// String 返回节点类型的字符串表示。
func (nt NodeType) String() string {
	return string(nt)
}

// String returns a string representation of the execution mode.
// String 返回执行模式的字符串表示。
func (em ExecutionMode) String() string {
	return string(em)
}

// String returns a string representation of the failure mode.
// String 返回失败模式的字符串表示。
func (fm FailureMode) String() string {
	return string(fm)
}

// Error returns the error message.
// Error 返回错误消息。
func (ve ValidationError) Error() string {
	if ve.NodeID != "" {
		return fmt.Sprintf("[%s] %s (node: %s)", ve.Code, ve.Message, ve.NodeID)
	}
	return fmt.Sprintf("[%s] %s", ve.Code, ve.Message)
}
