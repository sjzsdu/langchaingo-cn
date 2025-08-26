// Package graph - Middleware implementation
// 包 graph - 中间件实现
package graph

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ================================
// Built-in Middleware 内置中间件
// ================================

// LoggingMiddleware provides logging functionality for node execution.
// LoggingMiddleware 为节点执行提供日志功能。
type LoggingMiddleware struct {
	// Logger is the logger to use. If nil, uses default log package.
	Logger interface {
		Printf(format string, v ...interface{})
	}

	// LogLevel specifies what to log.
	LogLevel LogLevel

	// IncludeState specifies whether to include state information in logs.
	IncludeState bool
}

// LogLevel represents the logging level.
// LogLevel 表示日志级别。
type LogLevel int

const (
	// LogLevelNone disables logging.
	LogLevelNone LogLevel = iota
	// LogLevelError logs only errors.
	LogLevelError
	// LogLevelWarn logs warnings and errors.
	LogLevelWarn
	// LogLevelInfo logs info, warnings, and errors.
	LogLevelInfo
	// LogLevelDebug logs everything.
	LogLevelDebug
)

// NewLoggingMiddleware creates a new logging middleware.
// NewLoggingMiddleware 创建一个新的日志中间件。
func NewLoggingMiddleware(level LogLevel) *LoggingMiddleware {
	return &LoggingMiddleware{
		LogLevel:     level,
		IncludeState: false,
	}
}

// Process implements the Middleware interface.
// Process 实现 Middleware 接口。
func (lm *LoggingMiddleware) Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error) {
	if lm.LogLevel == LogLevelNone {
		return next(ctx, state)
	}

	logger := lm.Logger
	if logger == nil {
		logger = log.Default()
	}

	// Log start
	if lm.LogLevel >= LogLevelDebug {
		logger.Printf("Starting node execution - Current Node: %s", state.CurrentNode)
		if lm.IncludeState {
			logger.Printf("State: %+v", state)
		}
	}

	start := time.Now()
	result, err := next(ctx, state)
	duration := time.Since(start)

	// Log result
	if err != nil && lm.LogLevel >= LogLevelError {
		logger.Printf("Node execution failed - Node: %s, Error: %v, Duration: %v", state.CurrentNode, err, duration)
	} else if lm.LogLevel >= LogLevelInfo {
		logger.Printf("Node execution completed - Node: %s, Duration: %v", state.CurrentNode, duration)
	}

	return result, err
}

// ================================
// Metrics Middleware 指标中间件
// ================================

// MetricsMiddleware collects execution metrics.
// MetricsMiddleware 收集执行指标。
type MetricsMiddleware struct {
	// metrics stores the collected metrics.
	metrics map[string]*NodeMetrics

	// lock protects concurrent access to metrics.
	lock sync.RWMutex
}

// NodeMetrics contains metrics for a specific node.
// NodeMetrics 包含特定节点的指标。
type NodeMetrics struct {
	// ExecutionCount is the total number of executions.
	ExecutionCount int64 `json:"execution_count"`

	// SuccessCount is the number of successful executions.
	SuccessCount int64 `json:"success_count"`

	// ErrorCount is the number of failed executions.
	ErrorCount int64 `json:"error_count"`

	// TotalDuration is the total execution time.
	TotalDuration time.Duration `json:"total_duration"`

	// MinDuration is the minimum execution time.
	MinDuration time.Duration `json:"min_duration"`

	// MaxDuration is the maximum execution time.
	MaxDuration time.Duration `json:"max_duration"`

	// LastExecution is the timestamp of the last execution.
	LastExecution time.Time `json:"last_execution"`

	// LastError is the last error that occurred.
	LastError string `json:"last_error,omitempty"`
}

// NewMetricsMiddleware creates a new metrics middleware.
// NewMetricsMiddleware 创建一个新的指标中间件。
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: make(map[string]*NodeMetrics),
	}
}

// Process implements the Middleware interface.
// Process 实现 Middleware 接口。
func (mm *MetricsMiddleware) Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error) {
	// Try to get nodeID from state, fallback to "unknown" if not available
	nodeID := state.CurrentNode
	if nodeID == "" {
		nodeID = "unknown"
	}
	
	start := time.Now()

	result, err := next(ctx, state)
	duration := time.Since(start)

	// Use the result state's CurrentNode if available
	if result != nil && result.CurrentNode != "" {
		nodeID = result.CurrentNode
	}

	mm.recordMetrics(nodeID, duration, err)

	return result, err
}

// recordMetrics records metrics for a node execution.
// recordMetrics 记录节点执行的指标。
func (mm *MetricsMiddleware) recordMetrics(nodeID string, duration time.Duration, err error) {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	metrics, exists := mm.metrics[nodeID]
	if !exists {
		metrics = &NodeMetrics{
			MinDuration: duration,
			MaxDuration: duration,
		}
		mm.metrics[nodeID] = metrics
	}

	metrics.ExecutionCount++
	metrics.TotalDuration += duration
	metrics.LastExecution = time.Now()

	if err != nil {
		metrics.ErrorCount++
		metrics.LastError = err.Error()
	} else {
		metrics.SuccessCount++
		metrics.LastError = ""
	}

	if duration < metrics.MinDuration {
		metrics.MinDuration = duration
	}
	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}
}

// GetMetrics returns metrics for all nodes.
// GetMetrics 返回所有节点的指标。
func (mm *MetricsMiddleware) GetMetrics() map[string]*NodeMetrics {
	mm.lock.RLock()
	defer mm.lock.RUnlock()

	result := make(map[string]*NodeMetrics)
	for k, v := range mm.metrics {
		// Create a copy to avoid concurrent access issues
		result[k] = &NodeMetrics{
			ExecutionCount: v.ExecutionCount,
			SuccessCount:   v.SuccessCount,
			ErrorCount:     v.ErrorCount,
			TotalDuration:  v.TotalDuration,
			MinDuration:    v.MinDuration,
			MaxDuration:    v.MaxDuration,
			LastExecution:  v.LastExecution,
			LastError:      v.LastError,
		}
	}
	return result
}

// GetNodeMetrics returns metrics for a specific node.
// GetNodeMetrics 返回特定节点的指标。
func (mm *MetricsMiddleware) GetNodeMetrics(nodeID string) (*NodeMetrics, bool) {
	mm.lock.RLock()
	defer mm.lock.RUnlock()

	metrics, exists := mm.metrics[nodeID]
	if !exists {
		return nil, false
	}

	// Return a copy
	return &NodeMetrics{
		ExecutionCount: metrics.ExecutionCount,
		SuccessCount:   metrics.SuccessCount,
		ErrorCount:     metrics.ErrorCount,
		TotalDuration:  metrics.TotalDuration,
		MinDuration:    metrics.MinDuration,
		MaxDuration:    metrics.MaxDuration,
		LastExecution:  metrics.LastExecution,
		LastError:      metrics.LastError,
	}, true
}

// ResetMetrics resets all metrics.
// ResetMetrics 重置所有指标。
func (mm *MetricsMiddleware) ResetMetrics() {
	mm.lock.Lock()
	defer mm.lock.Unlock()
	mm.metrics = make(map[string]*NodeMetrics)
}

// ================================
// Timeout Middleware 超时中间件
// ================================

// TimeoutMiddleware enforces execution timeouts.
// TimeoutMiddleware 强制执行超时。
type TimeoutMiddleware struct {
	// DefaultTimeout is the default timeout for nodes that don't specify one.
	DefaultTimeout time.Duration

	// PerNodeTimeouts maps node IDs to their specific timeouts.
	PerNodeTimeouts map[string]time.Duration

	// lock protects concurrent access to per-node timeouts.
	lock sync.RWMutex
}

// NewTimeoutMiddleware creates a new timeout middleware.
// NewTimeoutMiddleware 创建一个新的超时中间件。
func NewTimeoutMiddleware(defaultTimeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		DefaultTimeout:  defaultTimeout,
		PerNodeTimeouts: make(map[string]time.Duration),
	}
}

// SetNodeTimeout sets a specific timeout for a node.
// SetNodeTimeout 为节点设置特定的超时时间。
func (tm *TimeoutMiddleware) SetNodeTimeout(nodeID string, timeout time.Duration) {
	tm.lock.Lock()
	defer tm.lock.Unlock()
	tm.PerNodeTimeouts[nodeID] = timeout
}

// Process implements the Middleware interface.
// Process 实现 Middleware 接口。
func (tm *TimeoutMiddleware) Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error) {
	timeout := tm.getTimeoutForNode(state.CurrentNode)
	if timeout <= 0 {
		return next(ctx, state)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return next(ctx, state)
}

// getTimeoutForNode gets the timeout for a specific node.
// getTimeoutForNode 获取特定节点的超时时间。
func (tm *TimeoutMiddleware) getTimeoutForNode(nodeID string) time.Duration {
	tm.lock.RLock()
	defer tm.lock.RUnlock()

	if timeout, exists := tm.PerNodeTimeouts[nodeID]; exists {
		return timeout
	}
	return tm.DefaultTimeout
}

// ================================
// Retry Middleware 重试中间件
// ================================

// RetryMiddleware provides retry functionality.
// RetryMiddleware 提供重试功能。
type RetryMiddleware struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// RetryDelay is the delay between retry attempts.
	RetryDelay time.Duration

	// BackoffMultiplier multiplies the delay after each retry.
	BackoffMultiplier float64

	// RetryableErrors is a list of error types that should trigger retries.
	RetryableErrors []string

	// ShouldRetry is a custom function to determine if an error should trigger a retry.
	ShouldRetry func(error) bool
}

// NewRetryMiddleware creates a new retry middleware.
// NewRetryMiddleware 创建一个新的重试中间件。
func NewRetryMiddleware(maxRetries int, retryDelay time.Duration) *RetryMiddleware {
	return &RetryMiddleware{
		MaxRetries:        maxRetries,
		RetryDelay:        retryDelay,
		BackoffMultiplier: 1.0,
		RetryableErrors:   []string{},
	}
}

// Process implements the Middleware interface.
// Process 实现 Middleware 接口。
func (rm *RetryMiddleware) Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error) {
	var lastErr error
	delay := rm.RetryDelay

	for attempt := 0; attempt <= rm.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			// Apply backoff
			delay = time.Duration(float64(delay) * rm.BackoffMultiplier)
		}

		result, err := next(ctx, state)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if we should retry this error
		if !rm.shouldRetryError(err) {
			break
		}
	}

	return nil, lastErr
}

// shouldRetryError determines if an error should trigger a retry.
// shouldRetryError 确定错误是否应该触发重试。
func (rm *RetryMiddleware) shouldRetryError(err error) bool {
	// Use custom function if provided
	if rm.ShouldRetry != nil {
		return rm.ShouldRetry(err)
	}

	// Check against retryable error types
	errStr := err.Error()
	for _, retryableErr := range rm.RetryableErrors {
		if errStr == retryableErr {
			return true
		}
	}

	// Don't retry context cancellation errors
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Default to retrying all other errors
	return true
}

// ================================
// Circuit Breaker Middleware 断路器中间件
// ================================

// CircuitBreakerMiddleware implements a circuit breaker pattern.
// CircuitBreakerMiddleware 实现断路器模式。
type CircuitBreakerMiddleware struct {
	// FailureThreshold is the number of failures that triggers the circuit breaker.
	FailureThreshold int

	// ResetTimeout is how long to wait before attempting to reset the circuit breaker.
	ResetTimeout time.Duration

	// state tracks the current state of the circuit breaker.
	state CircuitState

	// failures tracks the number of consecutive failures.
	failures int

	// lastFailureTime tracks when the last failure occurred.
	lastFailureTime time.Time

	// lock protects concurrent access.
	lock sync.RWMutex
}

// CircuitState represents the state of the circuit breaker.
// CircuitState 表示断路器的状态。
type CircuitState int

const (
	// CircuitStateClosed means the circuit is allowing requests through.
	CircuitStateClosed CircuitState = iota
	// CircuitStateOpen means the circuit is blocking requests.
	CircuitStateOpen
	// CircuitStateHalfOpen means the circuit is testing if requests should be allowed.
	CircuitStateHalfOpen
)

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware.
// NewCircuitBreakerMiddleware 创建一个新的断路器中间件。
func NewCircuitBreakerMiddleware(failureThreshold int, resetTimeout time.Duration) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		FailureThreshold: failureThreshold,
		ResetTimeout:     resetTimeout,
		state:            CircuitStateClosed,
	}
}

// Process implements the Middleware interface.
// Process 实现 Middleware 接口。
func (cb *CircuitBreakerMiddleware) Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error) {
	if !cb.allowRequest() {
		return nil, fmt.Errorf("circuit breaker is open")
	}

	result, err := next(ctx, state)
	cb.recordResult(err)

	return result, err
}

// allowRequest determines if a request should be allowed through.
// allowRequest 确定是否应该允许请求通过。
func (cb *CircuitBreakerMiddleware) allowRequest() bool {
	cb.lock.Lock()
	defer cb.lock.Unlock()

	switch cb.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) > cb.ResetTimeout {
			cb.state = CircuitStateHalfOpen
			return true
		}
		return false
	case CircuitStateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of a request.
// recordResult 记录请求的结果。
func (cb *CircuitBreakerMiddleware) recordResult(err error) {
	cb.lock.Lock()
	defer cb.lock.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()

		if cb.failures >= cb.FailureThreshold {
			cb.state = CircuitStateOpen
		}
	} else {
		// Success - reset failures and close circuit
		cb.failures = 0
		cb.state = CircuitStateClosed
	}
}

// GetState returns the current state of the circuit breaker.
// GetState 返回断路器的当前状态。
func (cb *CircuitBreakerMiddleware) GetState() CircuitState {
	cb.lock.RLock()
	defer cb.lock.RUnlock()
	return cb.state
}

// ================================
// Rate Limiting Middleware 限流中间件
// ================================

// RateLimitMiddleware implements rate limiting.
// RateLimitMiddleware 实现限流。
type RateLimitMiddleware struct {
	// Rate is the number of requests per second allowed.
	Rate float64

	// BurstSize is the maximum number of requests that can be made in a burst.
	BurstSize int

	// tokens tracks the current number of available tokens.
	tokens float64

	// lastRefill tracks when tokens were last refilled.
	lastRefill time.Time

	// lock protects concurrent access.
	lock sync.Mutex
}

// NewRateLimitMiddleware creates a new rate limiting middleware.
// NewRateLimitMiddleware 创建一个新的限流中间件。
func NewRateLimitMiddleware(rate float64, burstSize int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		Rate:       rate,
		BurstSize:  burstSize,
		tokens:     float64(burstSize),
		lastRefill: time.Now(),
	}
}

// Process implements the Middleware interface.
// Process 实现 Middleware 接口。
func (rl *RateLimitMiddleware) Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error) {
	if !rl.allowRequest() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	return next(ctx, state)
}

// allowRequest determines if a request should be allowed based on rate limiting.
// allowRequest 基于限流确定是否应该允许请求。
func (rl *RateLimitMiddleware) allowRequest() bool {
	rl.lock.Lock()
	defer rl.lock.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.lastRefill = now

	// Refill tokens
	rl.tokens += elapsed * rl.Rate
	if rl.tokens > float64(rl.BurstSize) {
		rl.tokens = float64(rl.BurstSize)
	}

	// Check if we have tokens available
	if rl.tokens >= 1.0 {
		rl.tokens--
		return true
	}

	return false
}

// ================================
// Validation Middleware 验证中间件
// ================================

// ValidationMiddleware validates state before and after node execution.
// ValidationMiddleware 在节点执行前后验证状态。
type ValidationMiddleware struct {
	// ValidateInput validates the input state.
	ValidateInput func(*State) error

	// ValidateOutput validates the output state.
	ValidateOutput func(*State) error

	// StrictMode determines if validation failures should stop execution.
	StrictMode bool
}

// NewValidationMiddleware creates a new validation middleware.
// NewValidationMiddleware 创建一个新的验证中间件。
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		StrictMode: true,
	}
}

// Process implements the Middleware interface.
// Process 实现 Middleware 接口。
func (vm *ValidationMiddleware) Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error) {
	// Validate input
	if vm.ValidateInput != nil {
		if err := vm.ValidateInput(state); err != nil {
			if vm.StrictMode {
				return nil, fmt.Errorf("input validation failed: %w", err)
			}
		}
	}

	result, err := next(ctx, state)
	if err != nil {
		return result, err
	}

	// Validate output
	if vm.ValidateOutput != nil {
		if err := vm.ValidateOutput(result); err != nil {
			if vm.StrictMode {
				return nil, fmt.Errorf("output validation failed: %w", err)
			}
		}
	}

	return result, nil
}

// ================================
// Middleware Chain 中间件链
// ================================

// MiddlewareChain represents a chain of middleware.
// MiddlewareChain 表示中间件链。
type MiddlewareChain struct {
	middleware []Middleware
}

// NewMiddlewareChain creates a new middleware chain.
// NewMiddlewareChain 创建一个新的中间件链。
func NewMiddlewareChain(middleware ...Middleware) *MiddlewareChain {
	return &MiddlewareChain{
		middleware: middleware,
	}
}

// Add adds middleware to the chain.
// Add 向链中添加中间件。
func (mc *MiddlewareChain) Add(middleware ...Middleware) *MiddlewareChain {
	mc.middleware = append(mc.middleware, middleware...)
	return mc
}

// Process applies all middleware in the chain.
// Process 应用链中的所有中间件。
func (mc *MiddlewareChain) Process(ctx context.Context, next func(ctx context.Context, state *State) (*State, error), state *State) (*State, error) {
	if len(mc.middleware) == 0 {
		return next(ctx, state)
	}

	// Build the middleware chain in reverse order
	finalFunc := next
	for i := len(mc.middleware) - 1; i >= 0; i-- {
		middleware := mc.middleware[i]
		prevFunc := finalFunc
		finalFunc = func(ctx context.Context, state *State) (*State, error) {
			return middleware.Process(ctx, prevFunc, state)
		}
	}

	return finalFunc(ctx, state)
}