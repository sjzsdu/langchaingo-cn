package main

import (
	"context"
	"fmt"
	"os"

	llmscn "github.com/sjzsdu/langchaingo-cn/llms"
	"github.com/tmc/langchaingo/llms"
)

func main() {
	// 从环境变量获取API密钥
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		fmt.Println("请设置LLM_API_KEY环境变量")
		return
	}

	// 创建DeepSeek LLM
	deepseekLLM, err := createDeepSeekLLM(apiKey)
	if err != nil {
		fmt.Printf("创建DeepSeek LLM失败: %v\n", err)
		return
	}

	// 创建Kimi LLM
	kimiLLM, err := createKimiLLM(apiKey)
	if err != nil {
		fmt.Printf("创建Kimi LLM失败: %v\n", err)
		return
	}

	// 创建Qwen LLM
	qwenLLM, err := createQwenLLM(apiKey)
	if err != nil {
		fmt.Printf("创建Qwen LLM失败: %v\n", err)
		return
	}

	// 创建Anthropic LLM（如果有API密钥）
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	var anthropicLLM llms.Model
	if anthropicAPIKey != "" {
		anthropicLLM, err = createAnthropicLLM(anthropicAPIKey)
		if err != nil {
			fmt.Printf("创建Anthropic LLM失败: %v\n", err)
		}
	}

	// 创建OpenAI LLM（如果有API密钥）
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	var openaiLLM llms.Model
	if openaiAPIKey != "" {
		openaiLLM, err = createOpenAILLM(openaiAPIKey)
		if err != nil {
			fmt.Printf("创建OpenAI LLM失败: %v\n", err)
		}
	}

	// 创建Ollama LLM（如果有Ollama服务器）
	var ollamaLLM llms.Model
	ollamaLLM, err = createOllamaLLM()
	if err != nil {
		fmt.Printf("创建Ollama LLM失败: %v\n", err)
	}

	// 使用DeepSeek LLM
	fmt.Println("\n=== 使用DeepSeek LLM ===")
	testLLM(deepseekLLM)

	// 使用Kimi LLM
	fmt.Println("\n=== 使用Kimi LLM ===")
	testLLM(kimiLLM)

	// 使用Qwen LLM
	fmt.Println("\n=== 使用Qwen LLM ===")
	testLLM(qwenLLM)

	// 使用Anthropic LLM（如果可用）
	if anthropicLLM != nil {
		fmt.Println("\n=== 使用Anthropic LLM ===")
		testLLM(anthropicLLM)
	}

	// 使用OpenAI LLM（如果可用）
	if openaiLLM != nil {
		fmt.Println("\n=== 使用OpenAI LLM ===")
		testLLM(openaiLLM)
	}

	// 使用Ollama LLM（如果可用）
	if ollamaLLM != nil {
		fmt.Println("\n=== 使用Ollama LLM ===")
		testLLM(ollamaLLM)
	}
}

// 创建DeepSeek LLM
func createDeepSeekLLM(apiKey string) (llms.Model, error) {
	return llmscn.CreateLLM(llmscn.DeepSeekLLM, map[string]interface{}{
		"api_key": apiKey,
		"model":   "deepseek-chat",
	})
}

// 创建Kimi LLM
func createKimiLLM(apiKey string) (llms.Model, error) {
	return llmscn.CreateLLM(llmscn.KimiLLM, map[string]interface{}{
		"api_key":      apiKey,
		"model":        "moonshot-v1-8k",
		"temperature":  0.7,
		"top_p":        0.9,
		"max_tokens":   1000,
	})
}

// 创建Qwen LLM
func createQwenLLM(apiKey string) (llms.Model, error) {
	return llmscn.CreateLLM(llmscn.QwenLLM, map[string]interface{}{
		"api_key":     apiKey,
		"model":       "qwen-turbo",
		"temperature": 0.8,
		"top_p":       0.95,
		"top_k":       50,
		"max_tokens":  2000,
	})
}

// 创建Anthropic LLM
func createAnthropicLLM(apiKey string) (llms.Model, error) {
	return llmscn.CreateLLM(llmscn.AnthropicLLM, map[string]interface{}{
		"api_key": apiKey,
		"model":   "claude-3-sonnet-20240229",
	})
}

// 创建OpenAI LLM
func createOpenAILLM(apiKey string) (llms.Model, error) {
	return llmscn.CreateLLM(llmscn.OpenAILLM, map[string]interface{}{
		"api_key": apiKey,
		"model":   "gpt-4",
	})
}

// 创建Ollama LLM
func createOllamaLLM() (llms.Model, error) {
	return llmscn.CreateLLM(llmscn.OllamaLLM, map[string]interface{}{
		"server_url": "http://localhost:11434",
		"model":      "llama3",
		"format":     "json",
		"system":     "你是一个有用的AI助手。",
	})
}

// 测试LLM
func testLLM(model llms.Model) {
	ctx := context.Background()

	// 测试简单文本生成
	prompt := "你好，请介绍一下自己"
	fmt.Printf("发送提示: %s\n", prompt)

	response, err := model.Call(ctx, prompt)
	if err != nil {
		fmt.Printf("调用LLM失败: %v\n", err)
		return
	}

	fmt.Printf("LLM回复: %s\n", response)

	// 测试多轮对话
	messages := []llms.ChatMessage{
		&llms.SystemChatMessage{
			Content: "你是一个有用的AI助手。",
		},
		&llms.HumanChatMessage{
			Content: "Go语言有哪些特点？",
		},
	}

	fmt.Println("发送多轮对话...")
	// 将 ChatMessage 转换为 MessageContent
	messageContents := make([]llms.MessageContent, len(messages))
	for i, msg := range messages {
		messageContents[i] = llms.MessageContent{
			Role:  msg.GetType(),
			Parts: []llms.ContentPart{llms.TextContent{Text: msg.GetContent()}},
		}
	}
	
	contentResponse, err := model.GenerateContent(ctx, messageContents)
	if err != nil {
		fmt.Printf("多轮对话失败: %v\n", err)
		return
	}

	if len(contentResponse.Choices) > 0 {
		fmt.Printf("多轮对话回复: %s\n", contentResponse.Choices[0].Content)
	}
}