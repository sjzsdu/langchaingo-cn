// Package graph - Node implementation
// 包 graph - 节点实现
package graph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// ================================
// Node Definition 节点定义
// ================================

// Node represents a node in the graph.
// Node 表示图中的一个节点。
type Node struct {
	// ID is the unique identifier for this node.
	ID string `json:"id"`

	// Name is a human-readable name for this node.
	Name string `json:"name"`

	// Type specifies the type of this node.
	Type NodeType `json:"type"`

	// Function is the main processing function for this node.
	Function NodeFunction `json:"-"`

	// ConditionFunc is used for condition nodes to determine the next path.
	ConditionFunc ConditionFunction `json:"-"`

	// Config contains configuration for this node.
	Config NodeConfig `json:"config"`

	// SubGraph contains a sub-graph for subgraph nodes.
	SubGraph *Graph `json:"-"`

	// Inputs defines the expected input parameters.
	Inputs []ParameterDef `json:"inputs,omitempty"`

	// Outputs defines the expected output parameters.
	Outputs []ParameterDef `json:"outputs,omitempty"`

	// Description describes what this node does.
	Description string `json:"description,omitempty"`

	// Version is the version of this node.
	Version string `json:"version,omitempty"`

	// Tags are labels for categorizing this node.
	Tags []string `json:"tags,omitempty"`

	// Middleware contains middleware specific to this node.
	middleware []Middleware

	// lock protects concurrent access to the node.
	lock sync.RWMutex
}

// ParameterDef defines an input or output parameter.
// ParameterDef 定义输入或输出参数。
type ParameterDef struct {
	// Name is the parameter name.
	Name string `json:"name"`

	// Type is the parameter type.
	Type string `json:"type"`

	// Required indicates if this parameter is required.
	Required bool `json:"required"`

	// Default is the default value for this parameter.
	Default interface{} `json:"default,omitempty"`

	// Description describes this parameter.
	Description string `json:"description,omitempty"`
}

// ================================
// Node Builder 节点构建器
// ================================

// NodeBuilder provides a fluent interface for building nodes.
// NodeBuilder 提供了构建节点的流畅接口。
type NodeBuilder struct {
	node *Node
}

// NewNode creates a new node builder.
// NewNode 创建一个新的节点构建器。
func NewNode(id string) *NodeBuilder {
	return &NodeBuilder{
		node: &Node{
			ID:         id,
			Type:       NodeTypeFunction,
			Config:     NodeConfig{},
			Inputs:     make([]ParameterDef, 0),
			Outputs:    make([]ParameterDef, 0),
			Tags:       make([]string, 0),
			middleware: make([]Middleware, 0),
		},
	}
}

// WithName sets the name of the node.
// WithName 设置节点的名称。
func (nb *NodeBuilder) WithName(name string) *NodeBuilder {
	nb.node.Name = name
	return nb
}

// WithType sets the type of the node.
// WithType 设置节点的类型。
func (nb *NodeBuilder) WithType(nodeType NodeType) *NodeBuilder {
	nb.node.Type = nodeType
	return nb
}

// WithFunction sets the processing function for the node.
// WithFunction 设置节点的处理函数。
func (nb *NodeBuilder) WithFunction(fn NodeFunction) *NodeBuilder {
	nb.node.Function = fn
	return nb
}

// WithCondition sets the condition function for condition nodes.
// WithCondition 设置条件节点的条件函数。
func (nb *NodeBuilder) WithCondition(fn ConditionFunction) *NodeBuilder {
	nb.node.ConditionFunc = fn
	nb.node.Type = NodeTypeCondition
	return nb
}

// WithSubGraph sets the sub-graph for subgraph nodes.
// WithSubGraph 设置子图节点的子图。
func (nb *NodeBuilder) WithSubGraph(subGraph *Graph) *NodeBuilder {
	nb.node.SubGraph = subGraph
	nb.node.Type = NodeTypeSubGraph
	return nb
}

// WithTimeout sets the timeout for the node.
// WithTimeout 设置节点的超时时间。
func (nb *NodeBuilder) WithTimeout(timeout time.Duration) *NodeBuilder {
	nb.node.Config.Timeout = timeout
	return nb
}

// WithRetries sets the retry configuration for the node.
// WithRetries 设置节点的重试配置。
func (nb *NodeBuilder) WithRetries(retries int, delay time.Duration) *NodeBuilder {
	nb.node.Config.Retries = retries
	nb.node.Config.RetryDelay = delay
	return nb
}

// WithFailureMode sets the failure handling mode for the node.
// WithFailureMode 设置节点的失败处理模式。
func (nb *NodeBuilder) WithFailureMode(mode FailureMode) *NodeBuilder {
	nb.node.Config.FailureMode = mode
	return nb
}

// WithDescription sets the description of the node.
// WithDescription 设置节点的描述。
func (nb *NodeBuilder) WithDescription(description string) *NodeBuilder {
	nb.node.Description = description
	return nb
}

// WithVersion sets the version of the node.
// WithVersion 设置节点的版本。
func (nb *NodeBuilder) WithVersion(version string) *NodeBuilder {
	nb.node.Version = version
	return nb
}

// WithTags adds tags to the node.
// WithTags 为节点添加标签。
func (nb *NodeBuilder) WithTags(tags ...string) *NodeBuilder {
	nb.node.Tags = append(nb.node.Tags, tags...)
	return nb
}

// WithInput adds an input parameter definition.
// WithInput 添加输入参数定义。
func (nb *NodeBuilder) WithInput(name, paramType string, required bool) *NodeBuilder {
	nb.node.Inputs = append(nb.node.Inputs, ParameterDef{
		Name:     name,
		Type:     paramType,
		Required: required,
	})
	return nb
}

// WithOutput adds an output parameter definition.
// WithOutput 添加输出参数定义。
func (nb *NodeBuilder) WithOutput(name, paramType string, description string) *NodeBuilder {
	nb.node.Outputs = append(nb.node.Outputs, ParameterDef{
		Name:        name,
		Type:        paramType,
		Description: description,
	})
	return nb
}

// WithMiddleware adds middleware to the node.
// WithMiddleware 为节点添加中间件。
func (nb *NodeBuilder) WithMiddleware(middleware ...Middleware) *NodeBuilder {
	nb.node.middleware = append(nb.node.middleware, middleware...)
	return nb
}

// WithMetadata sets metadata for the node.
// WithMetadata 设置节点的元数据。
func (nb *NodeBuilder) WithMetadata(key string, value interface{}) *NodeBuilder {
	if nb.node.Config.Metadata == nil {
		nb.node.Config.Metadata = make(map[string]interface{})
	}
	nb.node.Config.Metadata[key] = value
	return nb
}

// Build creates the node instance.
// Build 创建节点实例。
func (nb *NodeBuilder) Build() *Node {
	return nb.node
}

// ================================
// Node Execution 节点执行
// ================================

// Execute executes the node with the given context and state.
// Execute 使用给定的上下文和状态执行节点。
func (n *Node) Execute(ctx context.Context, state *State) (*State, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	// Record execution start
	step := ExecutionStep{
		NodeID:    n.ID,
		StartTime: time.Now(),
		Success:   false,
	}

	// Apply timeout if configured
	if n.Config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, n.Config.Timeout)
		defer cancel()
	}

	var result *State
	var err error

	// Execute with retries
	for attempt := 0; attempt <= n.Config.Retries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				err = ctx.Err()
				break
			case <-time.After(n.Config.RetryDelay):
			}
		}

		result, err = n.executeOnce(ctx, state)
		if err == nil {
			step.Success = true
			break
		}

		// Check if we should continue retrying
		if attempt < n.Config.Retries && n.shouldRetry(err) {
			continue
		}

		// Handle failure according to failure mode
		switch n.Config.FailureMode {
		case FailureModeContinue, FailureModeSkip:
			// Continue with current state, but mark as failed
			result = state
			err = nil
			step.Success = false
		case FailureModeStop:
			// Stop execution with error
			step.Success = false
		}
		break
	}

	// Record execution end
	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	if err != nil {
		step.Error = err.Error()
	}

	// Add execution step to history
	if result != nil {
		result.AddExecutionStep(step)
		result.CurrentNode = n.ID
	}

	return result, err
}

// executeOnce executes the node once without retries.
// executeOnce 执行节点一次，不重试。
func (n *Node) executeOnce(ctx context.Context, state *State) (*State, error) {
	// Apply middleware
	var finalFunc func(ctx context.Context, state *State) (*State, error)

	switch n.Type {
	case NodeTypeFunction:
		finalFunc = n.Function
	case NodeTypeCondition:
		finalFunc = n.executeCondition
	case NodeTypeParallel:
		finalFunc = n.executeParallel
	case NodeTypeLoop:
		finalFunc = n.executeLoop
	case NodeTypeSubGraph:
		finalFunc = n.executeSubGraph
	case NodeTypeStart:
		finalFunc = func(ctx context.Context, state *State) (*State, error) {
			return state, nil
		}
	case NodeTypeEnd:
		finalFunc = func(ctx context.Context, state *State) (*State, error) {
			return state, nil
		}
	default:
		return nil, fmt.Errorf("unsupported node type: %s", n.Type)
	}

	// Apply middleware in reverse order
	for i := len(n.middleware) - 1; i >= 0; i-- {
		middleware := n.middleware[i]
		prevFunc := finalFunc
		finalFunc = func(ctx context.Context, state *State) (*State, error) {
			return middleware.Process(ctx, prevFunc, state)
		}
	}

	return finalFunc(ctx, state)
}

// executeCondition executes a condition node.
// executeCondition 执行条件节点。
func (n *Node) executeCondition(ctx context.Context, state *State) (*State, error) {
	if n.ConditionFunc == nil {
		return nil, fmt.Errorf("condition function not set for condition node %s", n.ID)
	}

	nextNodeID, err := n.ConditionFunc(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("condition evaluation failed: %w", err)
	}

	// Store the next node ID in metadata for the router to use
	state.SetMetadata("next_node", nextNodeID)
	return state, nil
}

// executeParallel executes a parallel node (placeholder implementation).
// executeParallel 执行并行节点（占位符实现）。
func (n *Node) executeParallel(ctx context.Context, state *State) (*State, error) {
	// This would be implemented to execute multiple sub-nodes in parallel
	// For now, just return the state unchanged
	return state, nil
}

// executeLoop executes a loop node (placeholder implementation).
// executeLoop 执行循环节点（占位符实现）。
func (n *Node) executeLoop(ctx context.Context, state *State) (*State, error) {
	// This would be implemented to execute a sub-graph in a loop
	// For now, just return the state unchanged
	return state, nil
}

// executeSubGraph executes a sub-graph node.
// executeSubGraph 执行子图节点。
func (n *Node) executeSubGraph(ctx context.Context, state *State) (*State, error) {
	if n.SubGraph == nil {
		return nil, fmt.Errorf("sub-graph not set for subgraph node %s", n.ID)
	}

	runnable, err := n.SubGraph.Compile()
	if err != nil {
		return nil, fmt.Errorf("failed to compile sub-graph: %w", err)
	}

	return runnable.Invoke(ctx, state)
}

// shouldRetry determines if the node should retry on the given error.
// shouldRetry 确定节点是否应该在给定错误时重试。
func (n *Node) shouldRetry(err error) bool {
	// For now, retry on all errors except context cancellation
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}
	return true
}

// ================================
// Validation 验证
// ================================

// Validate validates the node configuration.
// Validate 验证节点配置。
func (n *Node) Validate() error {
	if n.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	switch n.Type {
	case NodeTypeFunction:
		if n.Function == nil {
			return fmt.Errorf("function node %s must have a function", n.ID)
		}
	case NodeTypeCondition:
		if n.ConditionFunc == nil {
			return fmt.Errorf("condition node %s must have a condition function", n.ID)
		}
	case NodeTypeSubGraph:
		if n.SubGraph == nil {
			return fmt.Errorf("subgraph node %s must have a sub-graph", n.ID)
		}
	}

	return nil
}

// ================================
// Utility Functions 工具函数
// ================================

// GetInputParameter gets an input parameter by name.
// GetInputParameter 根据名称获取输入参数。
func (n *Node) GetInputParameter(name string) (*ParameterDef, bool) {
	for _, param := range n.Inputs {
		if param.Name == name {
			return &param, true
		}
	}
	return nil, false
}

// GetOutputParameter gets an output parameter by name.
// GetOutputParameter 根据名称获取输出参数。
func (n *Node) GetOutputParameter(name string) (*ParameterDef, bool) {
	for _, param := range n.Outputs {
		if param.Name == name {
			return &param, true
		}
	}
	return nil, false
}

// HasTag checks if the node has a specific tag.
// HasTag 检查节点是否有特定标签。
func (n *Node) HasTag(tag string) bool {
	for _, t := range n.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Clone creates a copy of the node.
// Clone 创建节点的副本。
func (n *Node) Clone() *Node {
	n.lock.RLock()
	defer n.lock.RUnlock()

	clone := &Node{
		ID:            n.ID,
		Name:          n.Name,
		Type:          n.Type,
		Function:      n.Function,
		ConditionFunc: n.ConditionFunc,
		Config:        n.Config,
		SubGraph:      n.SubGraph, // Note: This is a shallow copy
		Inputs:        make([]ParameterDef, len(n.Inputs)),
		Outputs:       make([]ParameterDef, len(n.Outputs)),
		Description:   n.Description,
		Version:       n.Version,
		Tags:          make([]string, len(n.Tags)),
		middleware:    make([]Middleware, len(n.middleware)),
	}

	copy(clone.Inputs, n.Inputs)
	copy(clone.Outputs, n.Outputs)
	copy(clone.Tags, n.Tags)
	copy(clone.middleware, n.middleware)

	// Deep copy config metadata
	if n.Config.Metadata != nil {
		clone.Config.Metadata = make(map[string]interface{})
		for k, v := range n.Config.Metadata {
			clone.Config.Metadata[k] = v
		}
	}

	return clone
}

// ================================
// Pre-built Node Functions 预构建节点函数
// ================================

// MessageProcessorNode creates a node that processes messages.
// MessageProcessorNode 创建一个处理消息的节点。
func MessageProcessorNode(id string, processor func(ctx context.Context, messages []llms.MessageContent) ([]llms.MessageContent, error)) *Node {
	return NewNode(id).
		WithType(NodeTypeFunction).
		WithFunction(func(ctx context.Context, state *State) (*State, error) {
			processed, err := processor(ctx, state.Messages)
			if err != nil {
				return nil, err
			}
			state.Messages = processed
			return state, nil
		}).
		Build()
}

// VariableSetterNode creates a node that sets a variable.
// VariableSetterNode 创建一个设置变量的节点。
func VariableSetterNode(id, key string, value interface{}) *Node {
	return NewNode(id).
		WithType(NodeTypeFunction).
		WithFunction(func(ctx context.Context, state *State) (*State, error) {
			state.SetVariable(key, value)
			return state, nil
		}).
		Build()
}

// ConditionalNode creates a conditional node with a simple condition.
// ConditionalNode 创建一个带有简单条件的条件节点。
func ConditionalNode(id string, condition func(ctx context.Context, state *State) bool, trueNode, falseNode string) *Node {
	return NewNode(id).
		WithType(NodeTypeCondition).
		WithCondition(func(ctx context.Context, state *State) (string, error) {
			if condition(ctx, state) {
				return trueNode, nil
			}
			return falseNode, nil
		}).
		Build()
}
