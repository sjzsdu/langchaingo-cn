// Package graph - Graph implementation
// 包 graph - 图实现
package graph

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ================================
// Graph Definition 图定义
// ================================

// Graph represents a directed graph of nodes and edges.
// Graph 表示一个由节点和边组成的有向图。
type Graph struct {
	// ID is the unique identifier for this graph.
	ID string `json:"id"`

	// Name is a human-readable name for this graph.
	Name string `json:"name"`

	// Description describes what this graph does.
	Description string `json:"description"`

	// Version is the version of this graph.
	Version string `json:"version"`

	// Config contains configuration for this graph.
	Config GraphConfig `json:"config"`

	// nodes stores all nodes in the graph.
	nodes map[string]*Node

	// router handles edge routing.
	router *EdgeRouter

	// entryPoint is the ID of the entry point node.
	entryPoint string

	// middleware contains graph-level middleware.
	middleware []Middleware

	// stateManager handles state persistence.
	stateManager StateManager

	// lock protects concurrent access to the graph.
	lock sync.RWMutex
}

// ================================
// Graph Builder 图构建器
// ================================

// GraphBuilder provides a fluent interface for building graphs.
// GraphBuilder 提供了构建图的流畅接口。
type GraphBuilder struct {
	graph *Graph
}

// NewGraph creates a new graph builder.
// NewGraph 创建一个新的图构建器。
func NewGraph(id string) *GraphBuilder {
	return &GraphBuilder{
		graph: &Graph{
			ID:           id,
			nodes:        make(map[string]*Node),
			router:       NewEdgeRouter(),
			middleware:   make([]Middleware, 0),
			Config: GraphConfig{
				ExecutionMode:       ExecutionModeSequential,
				MaxConcurrency:      10,
				Timeout:             5 * time.Minute,
				EnableStateTracking: true,
			},
		},
	}
}

// WithName sets the name of the graph.
// WithName 设置图的名称。
func (gb *GraphBuilder) WithName(name string) *GraphBuilder {
	gb.graph.Name = name
	return gb
}

// WithDescription sets the description of the graph.
// WithDescription 设置图的描述。
func (gb *GraphBuilder) WithDescription(description string) *GraphBuilder {
	gb.graph.Description = description
	return gb
}

// WithVersion sets the version of the graph.
// WithVersion 设置图的版本。
func (gb *GraphBuilder) WithVersion(version string) *GraphBuilder {
	gb.graph.Version = version
	return gb
}

// WithConfig sets the configuration for the graph.
// WithConfig 设置图的配置。
func (gb *GraphBuilder) WithConfig(config GraphConfig) *GraphBuilder {
	gb.graph.Config = config
	return gb
}

// WithExecutionMode sets the execution mode for the graph.
// WithExecutionMode 设置图的执行模式。
func (gb *GraphBuilder) WithExecutionMode(mode ExecutionMode) *GraphBuilder {
	gb.graph.Config.ExecutionMode = mode
	return gb
}

// WithTimeout sets the timeout for the graph.
// WithTimeout 设置图的超时时间。
func (gb *GraphBuilder) WithTimeout(timeout time.Duration) *GraphBuilder {
	gb.graph.Config.Timeout = timeout
	return gb
}

// WithMaxConcurrency sets the maximum concurrency for the graph.
// WithMaxConcurrency 设置图的最大并发数。
func (gb *GraphBuilder) WithMaxConcurrency(maxConcurrency int) *GraphBuilder {
	gb.graph.Config.MaxConcurrency = maxConcurrency
	return gb
}

// WithStateTracking enables or disables state tracking.
// WithStateTracking 启用或禁用状态跟踪。
func (gb *GraphBuilder) WithStateTracking(enabled bool) *GraphBuilder {
	gb.graph.Config.EnableStateTracking = enabled
	return gb
}

// WithStateManager sets the state manager for the graph.
// WithStateManager 设置图的状态管理器。
func (gb *GraphBuilder) WithStateManager(manager StateManager) *GraphBuilder {
	gb.graph.stateManager = manager
	return gb
}

// WithMiddleware adds middleware to the graph.
// WithMiddleware 为图添加中间件。
func (gb *GraphBuilder) WithMiddleware(middleware ...Middleware) *GraphBuilder {
	gb.graph.middleware = append(gb.graph.middleware, middleware...)
	return gb
}

// WithMetadata sets metadata for the graph.
// WithMetadata 设置图的元数据。
func (gb *GraphBuilder) WithMetadata(key string, value interface{}) *GraphBuilder {
	if gb.graph.Config.Metadata == nil {
		gb.graph.Config.Metadata = make(map[string]interface{})
	}
	gb.graph.Config.Metadata[key] = value
	return gb
}

// AddNode adds a node to the graph.
// AddNode 向图添加节点。
func (gb *GraphBuilder) AddNode(node *Node) *GraphBuilder {
	if node != nil {
		gb.graph.nodes[node.ID] = node
	}
	return gb
}

// AddNodes adds multiple nodes to the graph.
// AddNodes 向图添加多个节点。
func (gb *GraphBuilder) AddNodes(nodes ...*Node) *GraphBuilder {
	for _, node := range nodes {
		if node != nil {
			gb.graph.nodes[node.ID] = node
		}
	}
	return gb
}

// AddEdge adds an edge to the graph.
// AddEdge 向图添加边。
func (gb *GraphBuilder) AddEdge(edge *Edge) *GraphBuilder {
	if edge != nil {
		gb.graph.router.AddEdge(*edge)
	}
	return gb
}

// AddEdges adds multiple edges to the graph.
// AddEdges 向图添加多个边。
func (gb *GraphBuilder) AddEdges(edges ...*Edge) *GraphBuilder {
	for _, edge := range edges {
		if edge != nil {
			gb.graph.router.AddEdge(*edge)
		}
	}
	return gb
}

// Connect creates a simple edge between two nodes.
// Connect 在两个节点之间创建简单的边。
func (gb *GraphBuilder) Connect(from, to string) *GraphBuilder {
	edge := NewEdge(fmt.Sprintf("%s_to_%s", from, to), from, to).Build()
	gb.graph.router.AddEdge(*edge)
	return gb
}

// ConnectWithCondition creates a conditional edge between two nodes.
// ConnectWithCondition 在两个节点之间创建条件边。
func (gb *GraphBuilder) ConnectWithCondition(from, to string, condition EdgeCondition) *GraphBuilder {
	edge := NewEdge(fmt.Sprintf("%s_to_%s_conditional", from, to), from, to).
		WithCondition(condition).
		Build()
	gb.graph.router.AddEdge(*edge)
	return gb
}

// SetEntryPoint sets the entry point of the graph.
// SetEntryPoint 设置图的入口点。
func (gb *GraphBuilder) SetEntryPoint(nodeID string) *GraphBuilder {
	gb.graph.entryPoint = nodeID
	return gb
}

// Build creates the graph instance.
// Build 创建图实例。
func (gb *GraphBuilder) Build() *Graph {
	return gb.graph
}

// ================================
// Graph Operations 图操作
// ================================

// GetNode returns a node by its ID.
// GetNode 根据ID返回节点。
func (g *Graph) GetNode(nodeID string) (*Node, bool) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	node, exists := g.nodes[nodeID]
	return node, exists
}

// GetNodes returns all nodes in the graph.
// GetNodes 返回图中的所有节点。
func (g *Graph) GetNodes() map[string]*Node {
	g.lock.RLock()
	defer g.lock.RUnlock()
	
	result := make(map[string]*Node)
	for id, node := range g.nodes {
		result[id] = node
	}
	return result
}

// AddNode adds a node to the graph.
// AddNode 向图添加节点。
func (g *Graph) AddNode(node *Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}
	
	if err := node.Validate(); err != nil {
		return fmt.Errorf("invalid node: %w", err)
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	
	g.nodes[node.ID] = node
	return nil
}

// RemoveNode removes a node from the graph.
// RemoveNode 从图中移除节点。
func (g *Graph) RemoveNode(nodeID string) bool {
	g.lock.Lock()
	defer g.lock.Unlock()
	
	if _, exists := g.nodes[nodeID]; !exists {
		return false
	}
	
	delete(g.nodes, nodeID)
	
	// Remove all edges connected to this node
	edges := g.router.GetEdgesFrom(nodeID)
	for _, edge := range edges {
		g.router.RemoveEdge(edge.ID)
	}
	
	edges = g.router.GetEdgesTo(nodeID)
	for _, edge := range edges {
		g.router.RemoveEdge(edge.ID)
	}
	
	return true
}

// AddEdge adds an edge to the graph.
// AddEdge 向图添加边。
func (g *Graph) AddEdge(edge *Edge) error {
	if edge == nil {
		return fmt.Errorf("edge cannot be nil")
	}
	
	if err := edge.Validate(); err != nil {
		return fmt.Errorf("invalid edge: %w", err)
	}
	
	// Check if nodes exist
	g.lock.RLock()
	_, fromExists := g.nodes[edge.From]
	_, toExists := g.nodes[edge.To]
	g.lock.RUnlock()
	
	if !fromExists {
		return fmt.Errorf("source node %s does not exist", edge.From)
	}
	if !toExists {
		return fmt.Errorf("destination node %s does not exist", edge.To)
	}
	
	g.router.AddEdge(*edge)
	return nil
}

// RemoveEdge removes an edge from the graph.
// RemoveEdge 从图移除边。
func (g *Graph) RemoveEdge(edgeID string) bool {
	return g.router.RemoveEdge(edgeID)
}

// GetEdgesFrom returns all edges from a specific node.
// GetEdgesFrom 返回来自特定节点的所有边。
func (g *Graph) GetEdgesFrom(nodeID string) []Edge {
	return g.router.GetEdgesFrom(nodeID)
}

// GetEdgesTo returns all edges to a specific node.
// GetEdgesTo 返回到特定节点的所有边。
func (g *Graph) GetEdgesTo(nodeID string) []Edge {
	return g.router.GetEdgesTo(nodeID)
}

// SetEntryPoint sets the entry point of the graph.
// SetEntryPoint 设置图的入口点。
func (g *Graph) SetEntryPoint(nodeID string) error {
	g.lock.RLock()
	_, exists := g.nodes[nodeID]
	g.lock.RUnlock()
	
	if !exists {
		return fmt.Errorf("entry point node %s does not exist", nodeID)
	}
	
	g.lock.Lock()
	g.entryPoint = nodeID
	g.lock.Unlock()
	
	return nil
}

// GetEntryPoint returns the entry point of the graph.
// GetEntryPoint 返回图的入口点。
func (g *Graph) GetEntryPoint() string {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.entryPoint
}

// ================================
// Graph Validation 图验证
// ================================

// Validate validates the entire graph.
// Validate 验证整个图。
func (g *Graph) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	// Check if entry point is set
	if g.entryPoint == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:    "NO_ENTRY_POINT",
			Message: "Graph has no entry point set",
		})
		result.Valid = false
	}

	// Check if entry point exists
	if g.entryPoint != "" {
		if _, exists := g.nodes[g.entryPoint]; !exists {
			result.Errors = append(result.Errors, ValidationError{
				Code:    "INVALID_ENTRY_POINT",
				Message: fmt.Sprintf("Entry point node %s does not exist", g.entryPoint),
				NodeID:  g.entryPoint,
			})
			result.Valid = false
		}
	}

	// Validate all nodes
	for _, node := range g.nodes {
		if err := node.Validate(); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Code:    "INVALID_NODE",
				Message: err.Error(),
				NodeID:  node.ID,
			})
			result.Valid = false
		}
	}

	// Validate edges
	edges := g.router.edges
	for _, edge := range edges {
		if err := edge.Validate(); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Code:    "INVALID_EDGE",
				Message: err.Error(),
				Details: map[string]interface{}{
					"edge_id": edge.ID,
					"from":    edge.From,
					"to":      edge.To,
				},
			})
			result.Valid = false
		}
	}

	// Check for unreachable nodes
	reachable := g.getReachableNodes()
	for nodeID := range g.nodes {
		if nodeID != g.entryPoint && !reachable[nodeID] {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    "UNREACHABLE_NODE",
				Message: fmt.Sprintf("Node %s is not reachable from entry point", nodeID),
				NodeID:  nodeID,
			})
		}
	}

	// Check for nodes with no outgoing edges (except END nodes)
	for nodeID, node := range g.nodes {
		if node.Type != NodeTypeEnd {
			edges := g.router.GetEdgesFrom(nodeID)
			if len(edges) == 0 {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Code:    "NO_OUTGOING_EDGES",
					Message: fmt.Sprintf("Node %s has no outgoing edges", nodeID),
					NodeID:  nodeID,
				})
			}
		}
	}

	return result
}

// getReachableNodes returns a set of nodes reachable from the entry point.
// getReachableNodes 返回从入口点可达的节点集合。
func (g *Graph) getReachableNodes() map[string]bool {
	reachable := make(map[string]bool)
	if g.entryPoint == "" {
		return reachable
	}

	visited := make(map[string]bool)
	queue := []string{g.entryPoint}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true
		reachable[current] = true

		// Add all connected nodes to queue
		edges := g.router.GetEdgesFrom(current)
		for _, edge := range edges {
			if !visited[edge.To] {
				queue = append(queue, edge.To)
			}
		}
	}

	return reachable
}

// ================================
// Graph Compilation 图编译
// ================================

// Compile compiles the graph into a runnable instance.
// Compile 将图编译为可运行的实例。
func (g *Graph) Compile() (*Runnable, error) {
	// Validate the graph
	validation := g.Validate()
	if !validation.Valid {
		var errorMessages []string
		for _, err := range validation.Errors {
			errorMessages = append(errorMessages, err.Error())
		}
		return nil, fmt.Errorf("graph validation failed: %s", strings.Join(errorMessages, "; "))
	}

	return &Runnable{
		graph: g,
	}, nil
}

// ================================
// Utility Functions 工具函数
// ================================

// Clone creates a deep copy of the graph.
// Clone 创建图的深拷贝。
func (g *Graph) Clone() *Graph {
	g.lock.RLock()
	defer g.lock.RUnlock()

	clone := &Graph{
		ID:           g.ID,
		Name:         g.Name,
		Description:  g.Description,
		Version:      g.Version,
		Config:       g.Config,
		nodes:        make(map[string]*Node),
		router:       NewEdgeRouter(),
		entryPoint:   g.entryPoint,
		middleware:   make([]Middleware, len(g.middleware)),
		stateManager: g.stateManager,
	}

	// Clone nodes
	for id, node := range g.nodes {
		clone.nodes[id] = node.Clone()
	}

	// Clone edges
	for _, edge := range g.router.edges {
		clone.router.AddEdge(*edge.Clone())
	}

	copy(clone.middleware, g.middleware)

	// Deep copy config metadata
	if g.Config.Metadata != nil {
		clone.Config.Metadata = make(map[string]interface{})
		for k, v := range g.Config.Metadata {
			clone.Config.Metadata[k] = v
		}
	}

	return clone
}

// GetNodeCount returns the number of nodes in the graph.
// GetNodeCount 返回图中节点的数量。
func (g *Graph) GetNodeCount() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return len(g.nodes)
}

// GetEdgeCount returns the number of edges in the graph.
// GetEdgeCount 返回图中边的数量。
func (g *Graph) GetEdgeCount() int {
	return len(g.router.edges)
}

// GetNodesByType returns all nodes of a specific type.
// GetNodesByType 返回特定类型的所有节点。
func (g *Graph) GetNodesByType(nodeType NodeType) []*Node {
	g.lock.RLock()
	defer g.lock.RUnlock()

	var result []*Node
	for _, node := range g.nodes {
		if node.Type == nodeType {
			result = append(result, node)
		}
	}
	return result
}

// GetNodesByTag returns all nodes with a specific tag.
// GetNodesByTag 返回具有特定标签的所有节点。
func (g *Graph) GetNodesByTag(tag string) []*Node {
	g.lock.RLock()
	defer g.lock.RUnlock()

	var result []*Node
	for _, node := range g.nodes {
		if node.HasTag(tag) {
			result = append(result, node)
		}
	}
	return result
}

// HasCycles checks if the graph has cycles.
// HasCycles 检查图是否有环。
func (g *Graph) HasCycles() bool {
	g.lock.RLock()
	defer g.lock.RUnlock()

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range g.nodes {
		if g.hasCyclesUtil(nodeID, visited, recStack) {
			return true
		}
	}
	return false
}

// hasCyclesUtil is a utility function for cycle detection using DFS.
// hasCyclesUtil 是使用DFS进行环检测的工具函数。
func (g *Graph) hasCyclesUtil(nodeID string, visited, recStack map[string]bool) bool {
	if recStack[nodeID] {
		return true // Back edge found, cycle detected
	}
	if visited[nodeID] {
		return false
	}

	visited[nodeID] = true
	recStack[nodeID] = true

	edges := g.router.GetEdgesFrom(nodeID)
	for _, edge := range edges {
		if g.hasCyclesUtil(edge.To, visited, recStack) {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}