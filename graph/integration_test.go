package graph_test

import (
	"context"
	"testing"
	"time"

	"github.com/sjzsdu/langchaingo-cn/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

// TestIntegrationBasicWorkflow tests a complete basic workflow
func TestIntegrationBasicWorkflow(t *testing.T) {
	// Create a simple workflow: input -> process -> output
	inputNode := graph.NewNode("input").
		WithName("Input Node").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "System initialized"},
				},
			})
			state.SetVariable("initialized", true)
			return state, nil
		}).
		Build()

	processNode := graph.NewNode("process").
		WithName("Process Node").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			// Simulate processing
			time.Sleep(10 * time.Millisecond)
			
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Processing completed"},
				},
			})
			state.SetVariable("processed", true)
			return state, nil
		}).
		WithTimeout(5 * time.Second).
		WithRetries(2, 1*time.Second).
		Build()

	outputNode := graph.NewNode("output").
		WithName("Output Node").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			state.SetMetadata("workflow_complete", true)
			state.SetMetadata("message_count", len(state.Messages))
			return state, nil
		}).
		Build()

	endNode := graph.NewNode("END").
		WithType(graph.NodeTypeEnd).
		Build()

	// Create the graph
	g := graph.NewGraph("integration_test").
		WithName("Integration Test Workflow").
		WithDescription("A complete workflow for integration testing").
		WithTimeout(30 * time.Second).
		AddNodes(inputNode, processNode, outputNode, endNode).
		Connect("input", "process").
		Connect("process", "output").
		Connect("output", "END").
		SetEntryPoint("input").
		Build()

	// Validate the graph
	validation := g.Validate()
	assert.True(t, validation.Valid, "Graph should be valid")
	assert.Empty(t, validation.Errors, "Should have no validation errors")

	// Compile the graph
	runnable, err := g.Compile()
	require.NoError(t, err, "Graph compilation should succeed")

	// Execute the graph
	state := graph.NewState("integration_test_001")
	state.AddMessage(llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: "Start workflow"},
		},
	})

	ctx := context.Background()
	result, err := runnable.Invoke(ctx, state)
	require.NoError(t, err, "Graph execution should succeed")
	require.NotNil(t, result, "Result should not be nil")

	// Verify the results
	assert.Equal(t, 3, len(result.Messages), "Should have 3 messages")
	assert.Equal(t, 3, len(result.History), "Should have 3 execution steps")

	// Check variables
	initialized, exists := result.GetVariable("initialized")
	assert.True(t, exists, "initialized variable should exist")
	assert.Equal(t, true, initialized, "Should be initialized")

	processed, exists := result.GetVariable("processed")
	assert.True(t, exists, "processed variable should exist")
	assert.Equal(t, true, processed, "Should be processed")

	// Check metadata
	complete, exists := result.GetMetadata("workflow_complete")
	assert.True(t, exists, "workflow_complete metadata should exist")
	assert.Equal(t, true, complete, "Workflow should be complete")

	messageCount, exists := result.GetMetadata("message_count")
	assert.True(t, exists, "message_count metadata should exist")
	assert.Equal(t, 3, messageCount, "Should have correct message count")

	// Verify execution history
	for i, step := range result.History {
		assert.True(t, step.Success, "Step %d should be successful", i)
		assert.NotEmpty(t, step.NodeID, "Step %d should have node ID", i)
		assert.Greater(t, step.Duration, time.Duration(0), "Step %d should have positive duration", i)
	}
}

// TestIntegrationConditionalWorkflow tests a workflow with conditional routing
func TestIntegrationConditionalWorkflow(t *testing.T) {
	// Create a workflow with conditional branching
	classifyNode := graph.NewNode("classify").
		WithName("Classifier").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			if len(state.Messages) > 0 {
				// Simple classification based on message content
				lastMsg := state.Messages[len(state.Messages)-1]
				for _, part := range lastMsg.Parts {
					if textPart, ok := part.(llms.TextContent); ok {
						if textPart.Text == "urgent" {
							state.SetVariable("priority", "high")
						} else {
							state.SetVariable("priority", "normal")
						}
						break
					}
				}
			}
			return state, nil
		}).
		Build()

	urgentNode := graph.NewNode("urgent_handler").
		WithName("Urgent Handler").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Urgent request processed immediately"},
				},
			})
			state.SetVariable("processed_urgently", true)
			return state, nil
		}).
		Build()

	normalNode := graph.NewNode("normal_handler").
		WithName("Normal Handler").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Normal request processed"},
				},
			})
			state.SetVariable("processed_normally", true)
			return state, nil
		}).
		Build()

	endNode := graph.NewNode("END").
		WithType(graph.NodeTypeEnd).
		Build()

	// Create conditional edges
	urgentEdge := graph.NewEdge("classify_to_urgent", "classify", "urgent_handler").
		WithCondition(func(ctx context.Context, state *graph.State) (bool, error) {
			priority, exists := state.GetVariable("priority")
			return exists && priority == "high", nil
		}).
		WithPriority(10).
		Build()

	normalEdge := graph.NewEdge("classify_to_normal", "classify", "normal_handler").
		AsDefault().
		Build()

	urgentToEnd := graph.AlwaysEdge("urgent_to_end", "urgent_handler", "END")
	normalToEnd := graph.AlwaysEdge("normal_to_end", "normal_handler", "END")

	// Build the graph
	g := graph.NewGraph("conditional_workflow").
		WithName("Conditional Workflow").
		AddNodes(classifyNode, urgentNode, normalNode, endNode).
		AddEdges(urgentEdge, normalEdge, urgentToEnd, normalToEnd).
		SetEntryPoint("classify").
		Build()

	runnable, err := g.Compile()
	require.NoError(t, err)

	// Test urgent path
	urgentState := graph.NewState("urgent_test")
	urgentState.AddMessage(llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: "urgent"},
		},
	})

	ctx := context.Background()
	result, err := runnable.Invoke(ctx, urgentState)
	require.NoError(t, err)

	// Verify urgent processing
	processed, exists := result.GetVariable("processed_urgently")
	assert.True(t, exists)
	assert.Equal(t, true, processed)

	_, exists = result.GetVariable("processed_normally")
	assert.False(t, exists, "Should not have processed normally")

	// Test normal path
	normalState := graph.NewState("normal_test")
	normalState.AddMessage(llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: "hello"},
		},
	})

	result, err = runnable.Invoke(ctx, normalState)
	require.NoError(t, err)

	// Verify normal processing
	processed, exists = result.GetVariable("processed_normally")
	assert.True(t, exists)
	assert.Equal(t, true, processed)

	_, exists = result.GetVariable("processed_urgently")
	assert.False(t, exists, "Should not have processed urgently")
}

// TestIntegrationWithMiddleware tests a workflow with middleware
func TestIntegrationWithMiddleware(t *testing.T) {
	// Create middleware
	loggingMiddleware := graph.NewLoggingMiddleware(graph.LogLevelInfo)
	metricsMiddleware := graph.NewMetricsMiddleware()
	timeoutMiddleware := graph.NewTimeoutMiddleware(10 * time.Second)

	// Create a node with middleware
	workNode := graph.NewNode("work").
		WithName("Work Node").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			// Simulate work
			time.Sleep(5 * time.Millisecond)
			state.SetVariable("work_done", true)
			return state, nil
		}).
		WithMiddleware(loggingMiddleware, metricsMiddleware).
		Build()

	endNode := graph.NewNode("END").
		WithType(graph.NodeTypeEnd).
		Build()

	// Create graph with graph-level middleware
	g := graph.NewGraph("middleware_test").
		WithName("Middleware Test").
		WithMiddleware(timeoutMiddleware).
		AddNodes(workNode, endNode).
		Connect("work", "END").
		SetEntryPoint("work").
		Build()

	runnable, err := g.Compile()
	require.NoError(t, err)

	// Execute multiple times to test metrics
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		state := graph.NewState("middleware_test")
		result, err := runnable.Invoke(ctx, state)
		require.NoError(t, err)

		workDone, exists := result.GetVariable("work_done")
		assert.True(t, exists)
		assert.Equal(t, true, workDone)
	}

	// Check metrics
	metrics := metricsMiddleware.GetMetrics()
	assert.NotEmpty(t, metrics)

	// The node ID might be "work" (after execution) or "unknown" (during execution)
	// Let's check both possibilities
	workMetrics, exists := metricsMiddleware.GetNodeMetrics("work")
	if !exists {
		workMetrics, exists = metricsMiddleware.GetNodeMetrics("unknown")
	}
	assert.True(t, exists)
	assert.Equal(t, int64(3), workMetrics.ExecutionCount)
	assert.Equal(t, int64(3), workMetrics.SuccessCount)
	assert.Equal(t, int64(0), workMetrics.ErrorCount)
	assert.Greater(t, workMetrics.TotalDuration, time.Duration(0))
}

// TestIntegrationStateManagement tests state persistence
func TestIntegrationStateManagement(t *testing.T) {
	// Create state manager
	stateManager := graph.NewMemoryStateManager(100)

	// Create a simple graph
	node := graph.NewNode("stateful_node").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			state.SetVariable("step", "completed")
			return state, nil
		}).
		Build()

	endNode := graph.NewNode("END").
		WithType(graph.NodeTypeEnd).
		Build()

	g := graph.NewGraph("stateful_graph").
		WithStateManager(stateManager).
		AddNodes(node, endNode).
		Connect("stateful_node", "END").
		SetEntryPoint("stateful_node").
		Build()

	runnable, err := g.Compile()
	require.NoError(t, err)

	// Execute and save state
	state := graph.NewState("persistent_test")
	ctx := context.Background()

	result, err := runnable.Invoke(ctx, state)
	require.NoError(t, err)

	// Save the result state
	err = stateManager.Save(ctx, result)
	require.NoError(t, err)

	// Load the state back
	loadedState, err := stateManager.Load(ctx, result.ID)
	require.NoError(t, err)
	assert.Equal(t, result.ID, loadedState.ID)

	step, exists := loadedState.GetVariable("step")
	assert.True(t, exists)
	assert.Equal(t, "completed", step)
}

// TestIntegrationComplexWorkflow tests a more complex real-world scenario
func TestIntegrationComplexWorkflow(t *testing.T) {
	// Create a complex workflow simulating a chat processing pipeline
	
	// Input validation
	validateNode := graph.NewNode("validate").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			if len(state.Messages) == 0 {
				state.SetVariable("validation_error", "No messages provided")
				return state, nil
			}
			state.SetVariable("validated", true)
			return state, nil
		}).
		Build()

	// Content analysis
	analyzeNode := graph.NewNode("analyze").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			// Simulate content analysis
			time.Sleep(2 * time.Millisecond)
			
			if len(state.Messages) > 0 {
				state.SetVariable("sentiment", "positive")
				state.SetVariable("topics", []string{"general"})
			}
			return state, nil
		}).
		Build()

	// Response generation
	generateNode := graph.NewNode("generate").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Thank you for your message!"},
				},
			})
			state.SetVariable("response_generated", true)
			return state, nil
		}).
		Build()

	// Error handling
	errorNode := graph.NewNode("error_handler").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Sorry, there was an error processing your request."},
				},
			})
			state.SetVariable("error_handled", true)
			return state, nil
		}).
		Build()

	endNode := graph.NewNode("END").WithType(graph.NodeTypeEnd).Build()

	// Conditional edges
	validateToAnalyze := graph.NewEdge("validate_to_analyze", "validate", "analyze").
		WithCondition(func(ctx context.Context, state *graph.State) (bool, error) {
			validated, exists := state.GetVariable("validated")
			return exists && validated == true, nil
		}).
		Build()

	validateToError := graph.NewEdge("validate_to_error", "validate", "error_handler").
		AsDefault().
		Build()

	analyzeToGenerate := graph.AlwaysEdge("analyze_to_generate", "analyze", "generate")
	generateToEnd := graph.AlwaysEdge("generate_to_end", "generate", "END")
	errorToEnd := graph.AlwaysEdge("error_to_end", "error_handler", "END")

	// Create middleware
	loggingMiddleware := graph.NewLoggingMiddleware(graph.LogLevelInfo)
	metricsMiddleware := graph.NewMetricsMiddleware()

	// Build the complex graph
	g := graph.NewGraph("complex_workflow").
		WithName("Complex Chat Processing Workflow").
		WithDescription("A comprehensive chat processing pipeline").
		WithVersion("1.0.0").
		WithTimeout(60 * time.Second).
		WithMiddleware(loggingMiddleware, metricsMiddleware).
		AddNodes(validateNode, analyzeNode, generateNode, errorNode, endNode).
		AddEdges(validateToAnalyze, validateToError, analyzeToGenerate, generateToEnd, errorToEnd).
		SetEntryPoint("validate").
		Build()

	// Validate the complex graph
	validation := g.Validate()
	assert.True(t, validation.Valid)
	assert.Empty(t, validation.Errors)

	runnable, err := g.Compile()
	require.NoError(t, err)

	// Test successful path
	successState := graph.NewState("complex_success")
	successState.AddMessage(llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: "Hello, how are you?"},
		},
	})

	ctx := context.Background()
	result, err := runnable.Invoke(ctx, successState)
	require.NoError(t, err)

	// Verify successful execution
	validated, exists := result.GetVariable("validated")
	assert.True(t, exists)
	assert.Equal(t, true, validated)

	sentiment, exists := result.GetVariable("sentiment")
	assert.True(t, exists)
	assert.Equal(t, "positive", sentiment)

	responseGenerated, exists := result.GetVariable("response_generated")
	assert.True(t, exists)
	assert.Equal(t, true, responseGenerated)

	assert.Equal(t, 2, len(result.Messages)) // Original + AI response

	// Test error path
	errorState := graph.NewState("complex_error")
	// Don't add any messages to trigger validation error

	result, err = runnable.Invoke(ctx, errorState)
	require.NoError(t, err)

	// Verify error handling
	validationError, exists := result.GetVariable("validation_error")
	assert.True(t, exists)
	assert.Equal(t, "No messages provided", validationError)

	errorHandled, exists := result.GetVariable("error_handled")
	assert.True(t, exists)
	assert.Equal(t, true, errorHandled)

	// Check metrics from multiple executions
	metrics := metricsMiddleware.GetMetrics()
	assert.NotEmpty(t, metrics)

	// Should have metrics for both successful and error paths
	validateMetrics, exists := metricsMiddleware.GetNodeMetrics("validate")
	assert.True(t, exists)
	assert.Equal(t, int64(2), validateMetrics.ExecutionCount) // Both test cases
}