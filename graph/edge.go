// Package graph - Edge implementation
// 包 graph - 边实现
package graph

import (
	"context"
	"fmt"
	"sync"
)

// ================================
// Edge Types 边类型
// ================================

// EdgeType represents the type of an edge.
// EdgeType 表示边的类型。
type EdgeType string

const (
	// EdgeTypeNormal represents a normal edge.
	EdgeTypeNormal EdgeType = "normal"
	// EdgeTypeConditional represents a conditional edge.
	EdgeTypeConditional EdgeType = "conditional"
	// EdgeTypePriority represents a priority-based edge.
	EdgeTypePriority EdgeType = "priority"
	// EdgeTypeDefault represents a default fallback edge.
	EdgeTypeDefault EdgeType = "default"
)

// ================================
// Edge Definition 边定义
// ================================

// Edge represents an edge in the graph.
// Edge 表示图中的一条边。
type Edge struct {
	// ID is the unique identifier for this edge.
	ID string `json:"id"`

	// Name is a human-readable name for this edge.
	Name string `json:"name,omitempty"`

	// Type specifies the type of this edge.
	Type EdgeType `json:"type"`

	// From is the ID of the source node.
	From string `json:"from"`

	// To is the ID of the destination node.
	To string `json:"to"`

	// Condition is the condition function for conditional edges.
	Condition EdgeCondition `json:"-"`

	// Priority is the priority of this edge (higher values have higher priority).
	Priority int `json:"priority,omitempty"`

	// Weight is the weight of this edge for weighted routing.
	Weight float64 `json:"weight,omitempty"`

	// Metadata contains custom metadata for this edge.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Description describes what this edge represents.
	Description string `json:"description,omitempty"`

	// Tags are labels for categorizing this edge.
	Tags []string `json:"tags,omitempty"`

	// Enabled indicates if this edge is currently enabled.
	Enabled bool `json:"enabled"`

	// lock protects concurrent access to the edge.
	lock sync.RWMutex
}

// EdgeCondition represents a condition function for conditional edges.
// EdgeCondition 表示条件边的条件函数。
type EdgeCondition func(ctx context.Context, state *State) (bool, error)

// ================================
// Edge Builder 边构建器
// ================================

// EdgeBuilder provides a fluent interface for building edges.
// EdgeBuilder 提供了构建边的流畅接口。
type EdgeBuilder struct {
	edge *Edge
}

// NewEdge creates a new edge builder.
// NewEdge 创建一个新的边构建器。
func NewEdge(id, from, to string) *EdgeBuilder {
	return &EdgeBuilder{
		edge: &Edge{
			ID:       id,
			Type:     EdgeTypeNormal,
			From:     from,
			To:       to,
			Priority: 0,
			Weight:   1.0,
			Metadata: make(map[string]interface{}),
			Tags:     make([]string, 0),
			Enabled:  true,
		},
	}
}

// WithName sets the name of the edge.
// WithName 设置边的名称。
func (eb *EdgeBuilder) WithName(name string) *EdgeBuilder {
	eb.edge.Name = name
	return eb
}

// WithType sets the type of the edge.
// WithType 设置边的类型。
func (eb *EdgeBuilder) WithType(edgeType EdgeType) *EdgeBuilder {
	eb.edge.Type = edgeType
	return eb
}

// WithCondition sets the condition function for conditional edges.
// WithCondition 设置条件边的条件函数。
func (eb *EdgeBuilder) WithCondition(condition EdgeCondition) *EdgeBuilder {
	eb.edge.Condition = condition
	eb.edge.Type = EdgeTypeConditional
	return eb
}

// WithPriority sets the priority of the edge.
// WithPriority 设置边的优先级。
func (eb *EdgeBuilder) WithPriority(priority int) *EdgeBuilder {
	eb.edge.Priority = priority
	if eb.edge.Type == EdgeTypeNormal {
		eb.edge.Type = EdgeTypePriority
	}
	return eb
}

// WithWeight sets the weight of the edge.
// WithWeight 设置边的权重。
func (eb *EdgeBuilder) WithWeight(weight float64) *EdgeBuilder {
	eb.edge.Weight = weight
	return eb
}

// WithDescription sets the description of the edge.
// WithDescription 设置边的描述。
func (eb *EdgeBuilder) WithDescription(description string) *EdgeBuilder {
	eb.edge.Description = description
	return eb
}

// WithTags adds tags to the edge.
// WithTags 为边添加标签。
func (eb *EdgeBuilder) WithTags(tags ...string) *EdgeBuilder {
	eb.edge.Tags = append(eb.edge.Tags, tags...)
	return eb
}

// WithMetadata sets metadata for the edge.
// WithMetadata 设置边的元数据。
func (eb *EdgeBuilder) WithMetadata(key string, value interface{}) *EdgeBuilder {
	eb.edge.Metadata[key] = value
	return eb
}

// WithEnabled sets whether the edge is enabled.
// WithEnabled 设置边是否启用。
func (eb *EdgeBuilder) WithEnabled(enabled bool) *EdgeBuilder {
	eb.edge.Enabled = enabled
	return eb
}

// AsDefault marks this edge as a default fallback edge.
// AsDefault 将此边标记为默认回退边。
func (eb *EdgeBuilder) AsDefault() *EdgeBuilder {
	eb.edge.Type = EdgeTypeDefault
	eb.edge.Priority = -1 // Default edges have lowest priority
	return eb
}

// Build creates the edge instance.
// Build 创建边实例。
func (eb *EdgeBuilder) Build() *Edge {
	return eb.edge
}

// ================================
// Edge Evaluation 边评估
// ================================

// CanTraverse checks if this edge can be traversed given the current state.
// CanTraverse 检查在当前状态下是否可以遍历此边。
func (e *Edge) CanTraverse(ctx context.Context, state *State) (bool, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if !e.Enabled {
		return false, nil
	}

	switch e.Type {
	case EdgeTypeNormal, EdgeTypePriority:
		return true, nil
	case EdgeTypeConditional:
		if e.Condition == nil {
			return false, fmt.Errorf("conditional edge %s has no condition function", e.ID)
		}
		return e.Condition(ctx, state)
	case EdgeTypeDefault:
		return true, nil
	default:
		return false, fmt.Errorf("unsupported edge type: %s", e.Type)
	}
}

// GetScore returns a score for this edge for routing decisions.
// GetScore 返回此边用于路由决策的分数。
func (e *Edge) GetScore(ctx context.Context, state *State) (float64, error) {
	canTraverse, err := e.CanTraverse(ctx, state)
	if err != nil {
		return 0, err
	}
	if !canTraverse {
		return 0, nil
	}

	// Base score is the weight
	score := e.Weight

	// Priority affects score (higher priority = higher score)
	score += float64(e.Priority) * 100

	// Default edges have very low score
	if e.Type == EdgeTypeDefault {
		score = 0.1
	}

	return score, nil
}

// ================================
// Edge Router 边路由器
// ================================

// EdgeRouter handles routing decisions between nodes.
// EdgeRouter 处理节点之间的路由决策。
type EdgeRouter struct {
	edges []Edge
	lock  sync.RWMutex
}

// NewEdgeRouter creates a new edge router.
// NewEdgeRouter 创建一个新的边路由器。
func NewEdgeRouter() *EdgeRouter {
	return &EdgeRouter{
		edges: make([]Edge, 0),
	}
}

// AddEdge adds an edge to the router.
// AddEdge 向路由器添加一条边。
func (er *EdgeRouter) AddEdge(edge Edge) {
	er.lock.Lock()
	defer er.lock.Unlock()
	er.edges = append(er.edges, edge)
}

// RemoveEdge removes an edge from the router.
// RemoveEdge 从路由器移除一条边。
func (er *EdgeRouter) RemoveEdge(edgeID string) bool {
	er.lock.Lock()
	defer er.lock.Unlock()

	for i, edge := range er.edges {
		if edge.ID == edgeID {
			er.edges = append(er.edges[:i], er.edges[i+1:]...)
			return true
		}
	}
	return false
}

// GetEdgesFrom returns all edges from a specific node.
// GetEdgesFrom 返回来自特定节点的所有边。
func (er *EdgeRouter) GetEdgesFrom(nodeID string) []Edge {
	er.lock.RLock()
	defer er.lock.RUnlock()

	var result []Edge
	for _, edge := range er.edges {
		if edge.From == nodeID {
			result = append(result, edge)
		}
	}
	return result
}

// GetEdgesTo returns all edges to a specific node.
// GetEdgesTo 返回到特定节点的所有边。
func (er *EdgeRouter) GetEdgesTo(nodeID string) []Edge {
	er.lock.RLock()
	defer er.lock.RUnlock()

	var result []Edge
	for _, edge := range er.edges {
		if edge.To == nodeID {
			result = append(result, edge)
		}
	}
	return result
}

// GetNextNode determines the next node to execute from a given node.
// GetNextNode 确定从给定节点执行的下一个节点。
func (er *EdgeRouter) GetNextNode(ctx context.Context, currentNodeID string, state *State) (string, error) {
	edges := er.GetEdgesFrom(currentNodeID)
	if len(edges) == 0 {
		return "", fmt.Errorf("no edges found from node %s", currentNodeID)
	}

	// Check if there's a specific next node set in metadata (for condition nodes)
	if nextNode, exists := state.GetMetadata("next_node"); exists {
		if nextNodeStr, ok := nextNode.(string); ok {
			// Verify that there's actually an edge to this node
			for _, edge := range edges {
				if edge.To == nextNodeStr {
					canTraverse, err := edge.CanTraverse(ctx, state)
					if err != nil {
						return "", err
					}
					if canTraverse {
						return nextNodeStr, nil
					}
				}
			}
		}
	}

	// Score all traversable edges
	type scoredEdge struct {
		edge  Edge
		score float64
	}

	var candidates []scoredEdge
	var defaultEdge *Edge

	for _, edge := range edges {
		score, err := edge.GetScore(ctx, state)
		if err != nil {
			return "", fmt.Errorf("error scoring edge %s: %w", edge.ID, err)
		}

		if score > 0 {
			candidates = append(candidates, scoredEdge{edge: edge, score: score})
		}

		// Keep track of default edge
		if edge.Type == EdgeTypeDefault {
			defaultEdge = &edge
		}
	}

	// If no candidates found, try default edge
	if len(candidates) == 0 {
		if defaultEdge != nil {
			canTraverse, err := defaultEdge.CanTraverse(ctx, state)
			if err != nil {
				return "", err
			}
			if canTraverse {
				return defaultEdge.To, nil
			}
		}
		return "", fmt.Errorf("no traversable edges found from node %s", currentNodeID)
	}

	// Sort candidates by score (highest first)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Return the highest scoring edge
	return candidates[0].edge.To, nil
}

// ================================
// Validation 验证
// ================================

// Validate validates the edge configuration.
// Validate 验证边配置。
func (e *Edge) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("edge ID cannot be empty")
	}
	if e.From == "" {
		return fmt.Errorf("edge %s must have a source node", e.ID)
	}
	if e.To == "" {
		return fmt.Errorf("edge %s must have a destination node", e.ID)
	}
	if e.Type == EdgeTypeConditional && e.Condition == nil {
		return fmt.Errorf("conditional edge %s must have a condition function", e.ID)
	}
	return nil
}

// ================================
// Utility Functions 工具函数
// ================================

// HasTag checks if the edge has a specific tag.
// HasTag 检查边是否有特定标签。
func (e *Edge) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Clone creates a copy of the edge.
// Clone 创建边的副本。
func (e *Edge) Clone() *Edge {
	e.lock.RLock()
	defer e.lock.RUnlock()

	clone := &Edge{
		ID:          e.ID,
		Name:        e.Name,
		Type:        e.Type,
		From:        e.From,
		To:          e.To,
		Condition:   e.Condition,
		Priority:    e.Priority,
		Weight:      e.Weight,
		Metadata:    make(map[string]interface{}),
		Description: e.Description,
		Tags:        make([]string, len(e.Tags)),
		Enabled:     e.Enabled,
	}

	// Deep copy metadata
	for k, v := range e.Metadata {
		clone.Metadata[k] = v
	}

	copy(clone.Tags, e.Tags)

	return clone
}

// Enable enables the edge.
// Enable 启用边。
func (e *Edge) Enable() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Enabled = true
}

// Disable disables the edge.
// Disable 禁用边。
func (e *Edge) Disable() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Enabled = false
}

// SetPriority sets the priority of the edge.
// SetPriority 设置边的优先级。
func (e *Edge) SetPriority(priority int) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Priority = priority
}

// SetWeight sets the weight of the edge.
// SetWeight 设置边的权重。
func (e *Edge) SetWeight(weight float64) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Weight = weight
}

// ================================
// Pre-built Edge Functions 预构建边函数
// ================================

// VariableConditionEdge creates a conditional edge based on a state variable.
// VariableConditionEdge 创建基于状态变量的条件边。
func VariableConditionEdge(id, from, to, variable string, expectedValue interface{}) *Edge {
	return NewEdge(id, from, to).
		WithCondition(func(ctx context.Context, state *State) (bool, error) {
			value, exists := state.GetVariable(variable)
			if !exists {
				return false, nil
			}
			return value == expectedValue, nil
		}).
		Build()
}

// MessageCountConditionEdge creates a conditional edge based on message count.
// MessageCountConditionEdge 创建基于消息数量的条件边。
func MessageCountConditionEdge(id, from, to string, minCount int) *Edge {
	return NewEdge(id, from, to).
		WithCondition(func(ctx context.Context, state *State) (bool, error) {
			return len(state.Messages) >= minCount, nil
		}).
		Build()
}

// AlwaysEdge creates an edge that is always traversable.
// AlwaysEdge 创建一条始终可遍历的边。
func AlwaysEdge(id, from, to string) *Edge {
	return NewEdge(id, from, to).Build()
}

// NeverEdge creates an edge that is never traversable.
// NeverEdge 创建一条永不可遍历的边。
func NeverEdge(id, from, to string) *Edge {
	return NewEdge(id, from, to).
		WithCondition(func(ctx context.Context, state *State) (bool, error) {
			return false, nil
		}).
		Build()
}