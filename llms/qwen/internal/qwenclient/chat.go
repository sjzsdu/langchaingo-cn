// Package qwenclient 提供了通义千问API的客户端实现
package qwenclient

// ChatMessage 表示聊天消息
type ChatMessage struct {
	// Role 是消息的角色，可以是"user"、"assistant"或"system"
	Role string `json:"role"`
	
	// Content 是消息的内容
	Content string `json:"content,omitempty"`
	
	// ContentParts 是多模态内容部分的列表
	ContentParts []ContentPart `json:"content_parts,omitempty"`
}

// ContentPart 表示多模态内容的一部分
type ContentPart struct {
	// Type 是内容部分的类型，可以是"text"或"image"
	Type string `json:"type"`
	
	// Text 是文本内容
	Text string `json:"text,omitempty"`
	
	// ImageURL 是图像URL
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL 表示图像URL
type ImageURL struct {
	// URL 是图像的URL
	URL string `json:"url"`
}

// ChatRequest 表示聊天请求
type ChatRequest struct {
	// Model 是要使用的模型名称
	Model string `json:"model"`
	
	// Input 是请求的输入
	Input ChatRequestInput `json:"input"`
	
	// Parameters 是请求的参数
	Parameters ChatRequestParameters `json:"parameters,omitempty"`
}

// ChatRequestInput 表示聊天请求的输入
type ChatRequestInput struct {
	// Prompt 是单轮对话的提示
	Prompt string `json:"prompt,omitempty"`
	
	// Messages 是多轮对话的消息列表
	Messages []ChatMessage `json:"messages,omitempty"`
}

// ChatRequestParameters 表示聊天请求的参数
type ChatRequestParameters struct {
	// Temperature 控制随机性，值越高回复越随机
	Temperature float64 `json:"temperature,omitempty"`
	
	// TopP 控制词汇选择的多样性
	TopP float64 `json:"top_p,omitempty"`
	
	// TopK 控制每一步考虑的词汇数量
	TopK int `json:"top_k,omitempty"`
	
	// MaxTokens 是生成的最大令牌数
	MaxTokens int `json:"max_tokens,omitempty"`
	
	// IncrementalOutput 表示是否启用流式输出
	IncrementalOutput bool `json:"incremental_output,omitempty"`
	
	// Seed 是随机数生成器的种子
	Seed int64 `json:"seed,omitempty"`
	
	// Tools 是可用工具的列表
	Tools []Tool `json:"tools,omitempty"`
	
	// ToolChoice 控制工具的选择
	ToolChoice *ToolChoice `json:"tool_choice,omitempty"`
	
	// ResultFormat 控制结果的格式
	ResultFormat string `json:"result_format,omitempty"`
}

// Tool 表示一个工具
type Tool struct {
	// Type 是工具的类型，通常为"function"
	Type string `json:"type"`
	
	// Function 是函数定义
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition 表示函数定义
type FunctionDefinition struct {
	// Name 是函数的名称
	Name string `json:"name"`
	
	// Description 是函数的描述
	Description string `json:"description"`
	
	// Parameters 是函数的参数
	Parameters map[string]interface{} `json:"parameters"`
	
	// Required 是必需的参数列表
	Required []string `json:"required,omitempty"`
}

// ToolChoice 表示工具选择
type ToolChoice struct {
	// Type 是工具选择的类型，可以是"auto"或"function"
	Type string `json:"type"`
	
	// Function 是函数选择
	Function *FunctionChoice `json:"function,omitempty"`
}

// FunctionChoice 表示函数选择
type FunctionChoice struct {
	// Name 是函数的名称
	Name string `json:"name"`
}

// ChatResponse 表示聊天响应
type ChatResponse struct {
	// RequestID 是请求的ID
	RequestID string `json:"request_id"`
	
	// Output 是响应的输出
	Output ChatResponseOutput `json:"output"`
	
	// Usage 是令牌使用情况
	Usage *Usage `json:"usage,omitempty"`
}

// ChatResponseOutput 表示聊天响应的输出
type ChatResponseOutput struct {
	// Text 是生成的文本
	Text string `json:"text,omitempty"`
	
	// FinishReason 是生成结束的原因
	FinishReason string `json:"finish_reason,omitempty"`
	
	// ToolCalls 是工具调用的列表
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall 表示工具调用
type ToolCall struct {
	// ID 是工具调用的ID
	ID string `json:"id"`
	
	// Type 是工具调用的类型，通常为"function"
	Type string `json:"type"`
	
	// Function 是函数调用
	Function FunctionCall `json:"function"`
}

// FunctionCall 表示函数调用
type FunctionCall struct {
	// Name 是函数的名称
	Name string `json:"name"`
	
	// Arguments 是函数的参数
	Arguments string `json:"arguments"`
}

// Usage 表示令牌使用情况
type Usage struct {
	// InputTokens 是输入的令牌数
	InputTokens int `json:"input_tokens"`
	
	// OutputTokens 是输出的令牌数
	OutputTokens int `json:"output_tokens"`
	
	// TotalTokens 是总令牌数
	TotalTokens int `json:"total_tokens"`
}

// ChatResponseChunk 表示流式聊天响应的一个块
type ChatResponseChunk struct {
	// RequestID 是请求的ID
	RequestID string `json:"request_id"`
	
	// Output 是响应块的输出
	Output ChatResponseChunkOutput `json:"output"`
}

// ChatResponseChunkOutput 表示流式聊天响应块的输出
type ChatResponseChunkOutput struct {
	// Text 是生成的文本
	Text string `json:"text,omitempty"`
	
	// FinishReason 是生成结束的原因
	FinishReason string `json:"finish_reason,omitempty"`
	
	// ToolCalls 是工具调用的列表
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// OpenAIChatRequest 表示OpenAI兼容模式的聊天请求
type OpenAIChatRequest struct {
	// Model 是要使用的模型名称
	Model string `json:"model"`
	
	// Messages 是消息列表
	Messages []OpenAIChatMessage `json:"messages"`
	
	// Temperature 控制随机性
	Temperature float64 `json:"temperature,omitempty"`
	
	// MaxTokens 是生成的最大令牌数
	MaxTokens int `json:"max_tokens,omitempty"`
	
	// TopP 控制词汇选择的多样性
	TopP float64 `json:"top_p,omitempty"`
	
	// Stream 表示是否启用流式输出
	Stream bool `json:"stream,omitempty"`
	
	// Tools 是可用工具的列表
	Tools []OpenAITool `json:"tools,omitempty"`
	
	// ToolChoice 控制工具的选择
	ToolChoice interface{} `json:"tool_choice,omitempty"`
}

// OpenAIChatMessage 表示OpenAI兼容模式的聊天消息
type OpenAIChatMessage struct {
	// Role 是消息的角色
	Role string `json:"role"`
	
	// Content 是消息的内容，可以是字符串或内容部分的列表
	Content interface{} `json:"content"`
}

// OpenAIContentPart 表示OpenAI兼容模式的内容部分
type OpenAIContentPart struct {
	// Type 是内容部分的类型
	Type string `json:"type"`
	
	// Text 是文本内容
	Text string `json:"text,omitempty"`
	
	// ImageURL 是图像URL
	ImageURL *OpenAIImageURL `json:"image_url,omitempty"`
}

// OpenAIImageURL 表示OpenAI兼容模式的图像URL
type OpenAIImageURL struct {
	// URL 是图像的URL
	URL string `json:"url"`
}

// OpenAITool 表示OpenAI兼容模式的工具
type OpenAITool struct {
	// Type 是工具的类型
	Type string `json:"type"`
	
	// Function 是函数定义
	Function FunctionDefinition `json:"function"`
}

// OpenAIChatResponse 表示OpenAI兼容模式的聊天响应
type OpenAIChatResponse struct {
	// ID 是响应的ID
	ID string `json:"id"`
	
	// Object 是对象类型
	Object string `json:"object"`
	
	// Created 是创建时间
	Created int64 `json:"created"`
	
	// Model 是使用的模型
	Model string `json:"model"`
	
	// Choices 是选择列表
	Choices []OpenAIChatResponseChoice `json:"choices"`
	
	// Usage 是令牌使用情况
	Usage *OpenAIUsage `json:"usage,omitempty"`
}

// OpenAIChatResponseChoice 表示OpenAI兼容模式的响应选择
type OpenAIChatResponseChoice struct {
	// Index 是选择的索引
	Index int `json:"index"`
	
	// Message 是消息
	Message OpenAIChatResponseMessage `json:"message"`
	
	// FinishReason 是生成结束的原因
	FinishReason string `json:"finish_reason"`
}

// OpenAIChatResponseMessage 表示OpenAI兼容模式的响应消息
type OpenAIChatResponseMessage struct {
	// Role 是消息的角色
	Role string `json:"role"`
	
	// Content 是消息的内容
	Content string `json:"content"`
	
	// ToolCalls 是工具调用的列表
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}

// OpenAIToolCall 表示OpenAI兼容模式的工具调用
type OpenAIToolCall struct {
	// ID 是工具调用的ID
	ID string `json:"id"`
	
	// Type 是工具调用的类型
	Type string `json:"type"`
	
	// Function 是函数调用
	Function OpenAIFunctionCall `json:"function"`
}

// OpenAIFunctionCall 表示OpenAI兼容模式的函数调用
type OpenAIFunctionCall struct {
	// Name 是函数的名称
	Name string `json:"name"`
	
	// Arguments 是函数的参数
	Arguments string `json:"arguments"`
}

// OpenAIUsage 表示OpenAI兼容模式的令牌使用情况
type OpenAIUsage struct {
	// PromptTokens 是提示的令牌数
	PromptTokens int `json:"prompt_tokens"`
	
	// CompletionTokens 是补全的令牌数
	CompletionTokens int `json:"completion_tokens"`
	
	// TotalTokens 是总令牌数
	TotalTokens int `json:"total_tokens"`
}

// OpenAIChatResponseChunk 表示OpenAI兼容模式的流式响应块
type OpenAIChatResponseChunk struct {
	// ID 是响应的ID
	ID string `json:"id"`
	
	// Object 是对象类型
	Object string `json:"object"`
	
	// Created 是创建时间
	Created int64 `json:"created"`
	
	// Model 是使用的模型
	Model string `json:"model"`
	
	// Choices 是选择列表
	Choices []OpenAIChatResponseChunkChoice `json:"choices"`
}

// OpenAIChatResponseChunkChoice 表示OpenAI兼容模式的流式响应选择
type OpenAIChatResponseChunkChoice struct {
	// Index 是选择的索引
	Index int `json:"index"`
	
	// Delta 是增量内容
	Delta OpenAIChatResponseChunkDelta `json:"delta"`
	
	// FinishReason 是生成结束的原因
	FinishReason string `json:"finish_reason,omitempty"`
}

// OpenAIChatResponseChunkDelta 表示OpenAI兼容模式的流式响应增量
type OpenAIChatResponseChunkDelta struct {
	// Role 是消息的角色
	Role string `json:"role,omitempty"`
	
	// Content 是消息的内容
	Content string `json:"content,omitempty"`
	
	// ToolCalls 是工具调用的列表
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}