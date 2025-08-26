// Package examples demonstrates advanced features of the graph framework
// 包 examples 演示图框架的高级功能
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sjzsdu/langchaingo-cn/graph"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// Create an advanced workflow with conditional routing and middleware
	// 创建具有条件路由和中间件的高级工作流
	workflowGraph := createAdvancedWorkflowGraph()

	// Compile the graph
	// 编译图
	runnable, err := workflowGraph.Compile()
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	// Test different scenarios
	// 测试不同场景
	testScenarios := []struct {
		name    string
		message string
	}{
		{"Question", "What is the capital of France?"},
		{"Greeting", "Hello there!"},
		{"Task", "Please help me write a Go function"},
		{"Math", "Calculate 15 + 27"},
	}

	for _, scenario := range testScenarios {
		fmt.Printf("\n=== Testing Scenario: %s ===\n", scenario.name)
		
		// Create state for this scenario
		// 为此场景创建状态
		state := graph.NewState(fmt.Sprintf("scenario_%s_%d", scenario.name, time.Now().Unix()))
		state.AddMessage(llms.MessageContent{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: scenario.message},
			},
		})

		// Execute with timeout and tracing
		// 使用超时和追踪执行
		ctx := context.Background()
		result, err := runnable.InvokeWithOptions(ctx, state,
			graph.WithTimeout(30*time.Second),
			graph.WithTracing(true),
			graph.WithMaxSteps(10),
		)

		if err != nil {
			log.Printf("Failed to execute scenario %s: %v", scenario.name, err)
			continue
		}

		// Print results
		// 打印结果
		fmt.Printf("Messages: %d\n", len(result.Messages))
		for _, msg := range result.Messages {
			fmt.Printf("  [%s]: %s\n", msg.Role, getTextFromMessage(msg))
		}

		// Print metadata
		// 打印元数据
		fmt.Println("Metadata:")
		for key, value := range result.Metadata {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	// Show execution statistics
	// 显示执行统计
	fmt.Println("\n=== Execution Statistics ===")
	stats := runnable.GetExecutionStats()
	if stats != nil {
		fmt.Printf("Total Executions: %d\n", stats.TotalExecutions)
		fmt.Printf("Successful: %d\n", stats.SuccessfulExecutions)
		fmt.Printf("Failed: %d\n", stats.FailedExecutions)
		fmt.Printf("Average Duration: %v\n", stats.AverageExecutionTime)

		fmt.Println("\nNode Statistics:")
		for nodeID, count := range stats.NodeExecutionCount {
			duration := stats.NodeExecutionTime[nodeID]
			avgDuration := duration / time.Duration(count)
			fmt.Printf("  %s: %d executions, avg %v\n", nodeID, count, avgDuration)
		}
	}
}

// createAdvancedWorkflowGraph creates a workflow with conditional routing
// createAdvancedWorkflowGraph 创建具有条件路由的工作流
func createAdvancedWorkflowGraph() *graph.Graph {
	// Create middleware
	// 创建中间件
	loggingMiddleware := graph.NewLoggingMiddleware(graph.LogLevelInfo)
	metricsMiddleware := graph.NewMetricsMiddleware()
	timeoutMiddleware := graph.NewTimeoutMiddleware(15 * time.Second)

	// Input classification node
	// 输入分类节点
	classifyNode := graph.NewNode("classify_input").
		WithName("Input Classifier").
		WithDescription("Classifies the type of user input").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			if len(state.Messages) == 0 {
				return nil, fmt.Errorf("no messages to classify")
			}

			lastMessage := state.Messages[len(state.Messages)-1]
			text := strings.ToLower(getTextFromMessage(lastMessage))

			var category string
			if strings.Contains(text, "hello") || strings.Contains(text, "hi") || strings.Contains(text, "hey") {
				category = "greeting"
			} else if strings.Contains(text, "?") || strings.Contains(text, "what") || strings.Contains(text, "how") || strings.Contains(text, "why") {
				category = "question"
			} else if strings.Contains(text, "calculate") || strings.Contains(text, "+") || strings.Contains(text, "-") || strings.Contains(text, "*") || strings.Contains(text, "/") {
				category = "math"
			} else if strings.Contains(text, "help") || strings.Contains(text, "write") || strings.Contains(text, "create") {
				category = "task"
			} else {
				category = "general"
			}

			state.SetVariable("input_category", category)
			fmt.Printf("Classified input as: %s\n", category)
			return state, nil
		}).
		WithMiddleware(loggingMiddleware, metricsMiddleware).
		Build()

	// Greeting handler node
	// 问候处理节点
	greetingNode := graph.NewNode("handle_greeting").
		WithName("Greeting Handler").
		WithDescription("Handles greeting messages").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			responses := []string{
				"Hello! How can I help you today?",
				"Hi there! What would you like to know?",
				"Hey! I'm here to assist you.",
			}
			
			// Select a random response (simplified)
			response := responses[0]
			
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextPart(response),
				},
			})
			
			state.SetMetadata("response_type", "greeting")
			return state, nil
		}).
		WithTimeout(5 * time.Second).
		Build()

	// Question handler node
	// 问题处理节点
	questionNode := graph.NewNode("handle_question").
		WithName("Question Handler").
		WithDescription("Handles question messages").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			userMessage := state.Messages[len(state.Messages)-1]
			userText := getTextFromMessage(userMessage)
			
			// Simulate intelligent question answering
			// 模拟智能问答
			var response string
			if strings.Contains(strings.ToLower(userText), "capital") && strings.Contains(strings.ToLower(userText), "france") {
				response = "The capital of France is Paris."
			} else {
				response = fmt.Sprintf("That's an interesting question about '%s'. Let me think about that...", userText)
			}
			
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: response},
				},
			})
			
			state.SetMetadata("response_type", "question_answer")
			return state, nil
		}).
		WithTimeout(10 * time.Second).
		WithRetries(1, 2*time.Second).
		Build()

	// Math handler node
	// 数学处理节点
	mathNode := graph.NewNode("handle_math").
		WithName("Math Handler").
		WithDescription("Handles mathematical calculations").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			userMessage := state.Messages[len(state.Messages)-1]
			userText := getTextFromMessage(userMessage)
			
			// Simple math parsing (for demo purposes)
			// 简单数学解析（演示目的）
			var response string
			if strings.Contains(userText, "15 + 27") || strings.Contains(userText, "15+27") {
				response = "15 + 27 = 42"
			} else {
				response = "I can help with basic math operations. Please provide a clear mathematical expression."
			}
			
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: response},
				},
			})
			
			state.SetMetadata("response_type", "math_calculation")
			return state, nil
		}).
		Build()

	// Task handler node
	// 任务处理节点
	taskNode := graph.NewNode("handle_task").
		WithName("Task Handler").
		WithDescription("Handles task requests and coding help").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			userMessage := state.Messages[len(state.Messages)-1]
			userText := getTextFromMessage(userMessage)
			
			var response string
			if strings.Contains(strings.ToLower(userText), "go function") {
				response = `Here's a simple Go function example:

func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}

This function takes a name as input and returns a greeting message.`
			} else {
				response = "I'd be happy to help you with your task. Could you provide more specific details about what you need?"
			}
			
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: response},
				},
			})
			
			state.SetMetadata("response_type", "task_assistance")
			return state, nil
		}).
		WithTimeout(15 * time.Second).
		Build()

	// General handler node
	// 通用处理节点
	generalNode := graph.NewNode("handle_general").
		WithName("General Handler").
		WithDescription("Handles general messages").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			response := "I understand you have something to discuss. How can I help you today?"
			
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: response},
				},
			})
			
			state.SetMetadata("response_type", "general")
			return state, nil
		}).
		Build()

	// Finalization node
	// 最终化节点
	finalizeNode := graph.NewNode("finalize").
		WithName("Finalizer").
		WithDescription("Finalizes the response").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			state.SetMetadata("processing_complete", true)
			state.SetMetadata("final_message_count", len(state.Messages))
			state.SetMetadata("completion_time", time.Now())
			
			// Add conversation summary
			category, _ := state.GetVariable("input_category")
			state.SetMetadata("conversation_summary", fmt.Sprintf("Processed %s input with %d messages", category, len(state.Messages)))
			
			return state, nil
		}).
		Build()

	// End node
	// 结束节点
	endNode := graph.NewNode("END").
		WithType(graph.NodeTypeEnd).
		WithName("End").
		Build()

	// Create conditional edges based on input classification
	// 基于输入分类创建条件边
	greetingEdge := graph.NewEdge("classify_to_greeting", "classify_input", "handle_greeting").
		WithCondition(func(ctx context.Context, state *graph.State) (bool, error) {
			category, exists := state.GetVariable("input_category")
			return exists && category == "greeting", nil
		}).
		WithPriority(10).
		Build()

	questionEdge := graph.NewEdge("classify_to_question", "classify_input", "handle_question").
		WithCondition(func(ctx context.Context, state *graph.State) (bool, error) {
			category, exists := state.GetVariable("input_category")
			return exists && category == "question", nil
		}).
		WithPriority(10).
		Build()

	mathEdge := graph.NewEdge("classify_to_math", "classify_input", "handle_math").
		WithCondition(func(ctx context.Context, state *graph.State) (bool, error) {
			category, exists := state.GetVariable("input_category")
			return exists && category == "math", nil
		}).
		WithPriority(10).
		Build()

	taskEdge := graph.NewEdge("classify_to_task", "classify_input", "handle_task").
		WithCondition(func(ctx context.Context, state *graph.State) (bool, error) {
			category, exists := state.GetVariable("input_category")
			return exists && category == "task", nil
		}).
		WithPriority(10).
		Build()

	// Default edge to general handler
	// 默认边到通用处理器
	generalEdge := graph.NewEdge("classify_to_general", "classify_input", "handle_general").
		AsDefault().
		Build()

	// Edges from handlers to finalizer
	// 从处理器到最终化器的边
	greetingToFinalize := graph.AlwaysEdge("greeting_to_finalize", "handle_greeting", "finalize")
	questionToFinalize := graph.AlwaysEdge("question_to_finalize", "handle_question", "finalize")
	mathToFinalize := graph.AlwaysEdge("math_to_finalize", "handle_math", "finalize")
	taskToFinalize := graph.AlwaysEdge("task_to_finalize", "handle_task", "finalize")
	generalToFinalize := graph.AlwaysEdge("general_to_finalize", "handle_general", "finalize")

	// Edge from finalizer to end
	// 从最终化器到结束的边
	finalizeToEnd := graph.AlwaysEdge("finalize_to_end", "finalize", "END")

	// Build the graph with middleware
	// 使用中间件构建图
	return graph.NewGraph("advanced_workflow").
		WithName("Advanced Workflow").
		WithDescription("Advanced workflow with conditional routing and middleware").
		WithVersion("2.0.0").
		WithTimeout(60 * time.Second).
		WithExecutionMode(graph.ExecutionModeSequential).
		WithMiddleware(timeoutMiddleware).
		AddNodes(classifyNode, greetingNode, questionNode, mathNode, taskNode, generalNode, finalizeNode, endNode).
		AddEdges(greetingEdge, questionEdge, mathEdge, taskEdge, generalEdge).
		AddEdges(greetingToFinalize, questionToFinalize, mathToFinalize, taskToFinalize, generalToFinalize, finalizeToEnd).
		SetEntryPoint("classify_input").
		Build()
}

// getTextFromMessage extracts text from a message
// getTextFromMessage 从消息中提取文本
func getTextFromMessage(msg llms.MessageContent) string {
	if len(msg.Parts) == 0 {
		return ""
	}
	
	for _, part := range msg.Parts {
		if textPart, ok := part.(llms.TextContent); ok {
			return textPart.Text
		}
	}
	return ""
}