package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/sjzsdu/langchaingo-cn/llms/deepseek"
	"github.com/tmc/langchaingo/llms"
)

// 定义错误类型
var (
	ErrToolNotFound      = errors.New("工具未找到")
	ErrInvalidParameters = errors.New("无效的参数")
	ErrNoResponse        = errors.New("没有响应")
)

// 定义天气信息结构体
type WeatherInfo struct {
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
}

// 定义工具接口
type Tool interface {
	// 获取工具定义
	GetDefinition() llms.Tool
	// 执行工具
	Execute(args map[string]interface{}) (interface{}, error)
}

// 天气工具实现
type WeatherTool struct{}

// 获取工具定义
func (w *WeatherTool) GetDefinition() llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "获取指定位置和日期的天气信息",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "城市名称，如北京、上海等",
					},
					"date": map[string]interface{}{
						"type":        "string",
						"description": "查询日期，格式为YYYY-MM-DD",
					},
				},
				"required": []string{"location"},
			},
		},
	}
}

// 执行天气工具
func (w *WeatherTool) Execute(args map[string]interface{}) (interface{}, error) {
	// 解析参数
	location, ok := args["location"].(string)
	if !ok || location == "" {
		return nil, ErrInvalidParameters
	}

	// 获取日期参数，如果没有则使用明天的日期
	date, _ := args["date"].(string)
	if date == "" {
		tomorrow := time.Now().AddDate(0, 0, 1)
		date = tomorrow.Format("2006-01-02")
	}

	// 调用天气API获取数据
	return getWeather(location, date), nil
}

// 模拟天气API
func getWeather(_ string, _ string) WeatherInfo {
	// 这里只是模拟数据，实际应用中应该调用真实的天气API
	return WeatherInfo{
		Temperature: 23.5,
		Condition:   "晴朗",
		Humidity:    65,
	}
}

// 工具工厂，用于管理和注册工具
type ToolFactory struct {
	tools map[string]Tool
}

// 创建新的工具工厂
func NewToolFactory() *ToolFactory {
	return &ToolFactory{
		tools: make(map[string]Tool),
	}
}

// 注册工具
func (f *ToolFactory) RegisterTool(name string, tool Tool) {
	f.tools[name] = tool
}

// 获取工具定义列表
func (f *ToolFactory) GetToolDefinitions() []llms.Tool {
	definitions := make([]llms.Tool, 0, len(f.tools))
	for _, tool := range f.tools {
		definitions = append(definitions, tool.GetDefinition())
	}
	return definitions
}

// 执行工具
func (f *ToolFactory) ExecuteTool(name string, args map[string]interface{}) (interface{}, error) {
	tool, exists := f.tools[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}
	return tool.Execute(args)
}

// 聊天服务，用于处理聊天请求和工具调用
type ChatService struct {
	llm          *deepseek.LLM
	toolFactory  *ToolFactory
	systemPrompt string
}

// 创建新的聊天服务
func NewChatService(llm *deepseek.LLM, toolFactory *ToolFactory, systemPrompt string) *ChatService {
	return &ChatService{
		llm:          llm,
		toolFactory:  toolFactory,
		systemPrompt: systemPrompt,
	}
}

// HandleChat 处理聊天请求
func (s *ChatService) HandleChat(ctx context.Context, userPrompt string) (string, error) {
	// 创建聊天消息
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, s.systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
	}

	// 生成内容并处理工具调用
	completion, err := s.llm.GenerateContent(
		ctx,
		content,
		llms.WithMaxTokens(1000),
		llms.WithTemperature(0.7),
		llms.WithTools(s.toolFactory.GetToolDefinitions()),
	)
	if err != nil {
		return "", fmt.Errorf("生成内容失败: %w", err)
	}

	// 处理工具调用
	if len(completion.Choices) > 0 && len(completion.Choices[0].ToolCalls) > 0 {
		return s.handleToolCalls(ctx, userPrompt, completion.Choices[0].ToolCalls)
	}

	// 直接返回回复
	if len(completion.Choices) > 0 {
		return completion.Choices[0].Content, nil
	}

	return "", ErrNoResponse
}

// 处理工具调用
func (s *ChatService) handleToolCalls(ctx context.Context, userPrompt string, toolCalls []llms.ToolCall) (string, error) {
	fmt.Println("模型请求调用工具:")

	// 处理每个工具调用
	for _, toolCall := range toolCalls {
		fmt.Printf("工具: %s\n", toolCall.FunctionCall.Name)
		fmt.Printf("参数: %s\n", toolCall.FunctionCall.Arguments)

		// 解析参数
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
			return "", fmt.Errorf("解析参数失败: %w", err)
		}

		// 执行工具
		result, err := s.toolFactory.ExecuteTool(toolCall.FunctionCall.Name, args)
		if err != nil {
			return "", fmt.Errorf("执行工具失败: %w", err)
		}

		// 将结果转换为JSON
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("序列化结果失败: %w", err)
		}

		fmt.Printf("工具返回结果: %s\n\n", string(resultJSON))

		// 将工具调用结果发送回模型
		toolResult := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, s.systemPrompt),
			llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
			// 助手消息必须同时包含内容和工具调用
			{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					// 添加工具调用
					llms.ToolCall{
						ID:   toolCall.ID,
						Type: "function",
						FunctionCall: &llms.FunctionCall{
							Name:      toolCall.FunctionCall.Name,
							Arguments: toolCall.FunctionCall.Arguments,
						},
					},
				},
			},
			// 工具响应消息
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						Content:    string(resultJSON),
						ToolCallID: toolCall.ID,                // 使用工具调用的ID
						Name:       toolCall.FunctionCall.Name, // 使用工具的名称
					},
				},
			},
		}

		// 获取最终回复
		finalResponse, err := s.llm.GenerateContent(ctx, toolResult, llms.WithMaxTokens(1000))
		if err != nil {
			return "", fmt.Errorf("获取最终回复失败: %w", err)
		}

		// 返回最终回复
		if len(finalResponse.Choices) > 0 {
			return finalResponse.Choices[0].Content, nil
		}
	}

	return "", ErrNoResponse
}

func main() {
	// 初始化DeepSeek客户端
	llm, err := deepseek.New()
	if err != nil {
		log.Fatal("初始化DeepSeek客户端失败: ", err)
	}

	ctx := context.Background()

	// 创建工具工厂
	toolFactory := NewToolFactory()

	// 注册天气工具
	toolFactory.RegisterTool("get_weather", &WeatherTool{})

	// 创建聊天服务
	systemPrompt := "你是一个旅行助手，可以帮助用户查询天气信息并提供相应的旅行建议。"
	chatService := NewChatService(llm, toolFactory, systemPrompt)

	// 处理用户请求
	userPrompt := "我想知道北京明天的天气如何，我应该准备什么衣物？"
	response, err := chatService.HandleChat(ctx, userPrompt)
	if err != nil {
		log.Fatal("处理聊天请求失败: ", err)
	}

	// 输出回复
	fmt.Println("最终回复:")
	fmt.Println(response)
}
