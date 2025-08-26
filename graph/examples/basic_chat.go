// Package main demonstrates basic usage of the graph framework
// 包 main 演示图框架的基本用法
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sjzsdu/langchaingo-cn/graph"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// Create a basic chat graph
	// 创建基本聊天图
	chatGraph := createBasicChatGraph()

	// Compile the graph
	// 编译图
	runnable, err := chatGraph.Compile()
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	// Create initial state with a user message
	// 创建包含用户消息的初始状态
	state := graph.NewState("example_chat_001")
	state.AddMessage(llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: "Hello, how are you today?"},
		},
	})

	// Execute the graph
	// 执行图
	ctx := context.Background()
	result, err := runnable.Invoke(ctx, state)
	if err != nil {
		log.Fatalf("Failed to execute graph: %v", err)
	}

	// Print the result
	// 打印结果
	fmt.Println("Chat conversation:")
	for i, msg := range result.Messages {
		fmt.Printf("%d. [%s]: %s\n", i+1, msg.Role, getTextFromMessageBasic(msg))
	}

	// Print execution history
	// 打印执行历史
	fmt.Println("\nExecution History:")
	for i, step := range result.History {
		status := "SUCCESS"
		if step.Error != "" {
			status = "ERROR"
		}
		fmt.Printf("%d. Node: %s, Status: %s, Duration: %v\n",
			i+1, step.NodeID, status, step.Duration)
	}
}

// createBasicChatGraph creates a simple chat processing graph
// createBasicChatGraph 创建一个简单的聊天处理图
func createBasicChatGraph() *graph.Graph {
	// Create nodes
	// 创建节点

	// Input validation node
	// 输入验证节点
	validateNode := graph.NewNode("validate_input").
		WithName("Input Validator").
		WithDescription("Validates user input").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			if len(state.Messages) == 0 {
				return nil, fmt.Errorf("no messages provided")
			}

			lastMessage := state.Messages[len(state.Messages)-1]
			if lastMessage.Role != llms.ChatMessageTypeHuman {
				return nil, fmt.Errorf("last message must be from human")
			}

			fmt.Println("✓ Input validation passed")
			return state, nil
		}).
		WithTimeout(5 * time.Second).
		Build()

	// Content processing node
	// 内容处理节点
	processNode := graph.NewNode("process_content").
		WithName("Content Processor").
		WithDescription("Processes the user input and generates response").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			// Simulate processing time
			// 模拟处理时间
			time.Sleep(100 * time.Millisecond)

			// Get the last user message
			// 获取最后一条用户消息
			userMessage := state.Messages[len(state.Messages)-1]
			userText := getTextFromMessageBasic(userMessage)

			// Generate a simple response (in real scenario, this would call an LLM)
			// 生成简单回复（实际场景中，这里会调用LLM）
			var response string
			if userText == "Hello, how are you today?" {
				response = "Hello! I'm doing great, thank you for asking. How can I help you today?"
			} else {
				response = fmt.Sprintf("I received your message: '%s'. How can I assist you further?", userText)
			}

			// Add AI response to state
			// 向状态添加AI回复
			state.AddMessage(llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: response},
				},
			})

			fmt.Println("✓ Content processed and response generated")
			return state, nil
		}).
		WithTimeout(10*time.Second).
		WithRetries(2, 1*time.Second).
		Build()

	// Output formatting node
	// 输出格式化节点
	formatNode := graph.NewNode("format_output").
		WithName("Output Formatter").
		WithDescription("Formats the final output").
		WithFunction(func(ctx context.Context, state *graph.State) (*graph.State, error) {
			// Set metadata about the conversation
			// 设置对话的元数据
			state.SetMetadata("conversation_length", len(state.Messages))
			state.SetMetadata("last_updated", time.Now())
			state.SetMetadata("processing_complete", true)

			fmt.Println("✓ Output formatted")
			return state, nil
		}).
		Build()

	// End node
	// 结束节点
	endNode := graph.NewNode("END").
		WithType(graph.NodeTypeEnd).
		WithName("End").
		WithDescription("Marks the end of processing").
		Build()

	// Create edges
	// 创建边
	validateToProcess := graph.NewEdge("validate_to_process", "validate_input", "process_content").
		WithName("Validation to Processing").
		Build()

	processToFormat := graph.NewEdge("process_to_format", "process_content", "format_output").
		WithName("Processing to Formatting").
		Build()

	formatToEnd := graph.NewEdge("format_to_end", "format_output", "END").
		WithName("Formatting to End").
		Build()

	// Build the graph
	// 构建图
	return graph.NewGraph("basic_chat_graph").
		WithName("Basic Chat Graph").
		WithDescription("A simple chat processing workflow").
		WithVersion("1.0.0").
		WithTimeout(30*time.Second).
		AddNodes(validateNode, processNode, formatNode, endNode).
		AddEdges(validateToProcess, processToFormat, formatToEnd).
		SetEntryPoint("validate_input").
		Build()
}

// getTextFromMessageBasic extracts text from a message
// getTextFromMessageBasic 从消息中提取文本
func getTextFromMessageBasic(msg llms.MessageContent) string {
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
