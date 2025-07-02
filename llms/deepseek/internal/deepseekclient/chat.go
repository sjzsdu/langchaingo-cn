package deepseekclient

import (
	"context"
)

// ChatMessage represents a message in a chat conversation.
type ChatMessage struct {
	// Role is the role of the message sender (system, user, assistant, tool).
	Role string `json:"role"`
	// Content is the content of the message.
	Content string `json:"content,omitempty"`
	// ContentParts is the multi-modal content parts of the message.
	ContentParts []ContentPart `json:"content_parts,omitempty"`
	// ToolCallID is the ID of the tool call this message is responding to.
	ToolCallID string `json:"tool_call_id,omitempty"`
	// Name is the name of the tool that was called.
	Name string `json:"name,omitempty"`
	// Prefix is a flag to enable chat prefix completion (Beta).
	Prefix bool `json:"prefix,omitempty"`
	// ToolCalls is the list of tool calls in the message.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ContentPart represents a part of a multi-modal message content.
type ContentPart struct {
	// Type is the type of the content part (text, image_url).
	Type string `json:"type"`
	// Text is the text content.
	Text string `json:"text,omitempty"`
	// ImageURL is the image URL content.
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL content part.
type ImageURL struct {
	// URL is the URL of the image.
	URL string `json:"url"`
	// Detail is the detail level of the image (low, high).
	Detail string `json:"detail,omitempty"`
}

// ChatRequest represents a request to the DeepSeek chat completions API.
type ChatRequest struct {
	// Model is the model to use.
	Model string `json:"model"`
	// Messages is the list of messages in the conversation.
	Messages []ChatMessage `json:"messages"`
	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens int `json:"max_tokens,omitempty"`
	// Temperature is the sampling temperature.
	Temperature float64 `json:"temperature,omitempty"`
	// TopP is the nucleus sampling parameter.
	TopP float64 `json:"top_p,omitempty"`
	// N is the number of chat completion choices to generate.
	N int `json:"n,omitempty"`
	// Stop is a list of tokens at which to stop generation.
	Stop []string `json:"stop,omitempty"`
	// FrequencyPenalty penalizes repeated tokens.
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	// PresencePenalty penalizes repeated topics.
	PresencePenalty float64 `json:"presence_penalty,omitempty"`
	// Stream indicates whether to stream the response.
	Stream bool `json:"stream,omitempty"`
	// StreamingFunc is a function to be called for each chunk of a streaming response.
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
	// StreamingReasoningFunc is a function to be called for each chunk of a streaming reasoning response.
	StreamingReasoningFunc func(ctx context.Context, reasoningChunk, chunk []byte) error `json:"-"`
	// Tools is a list of tools available to the model.
	Tools []Tool `json:"tools,omitempty"`
	// ToolChoice controls which tool is used by the model.
	ToolChoice any `json:"tool_choice,omitempty"`
	// JSONMode enables JSON mode for structured output.
	JSONMode bool `json:"json_mode,omitempty"`
	// ResponseFormat specifies the format of the response.
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	// LogProbs enables token log probabilities.
	LogProbs bool `json:"logprobs,omitempty"`
	// TopLogProbs is the number of most likely tokens to return with log probabilities.
	TopLogProbs int `json:"top_logprobs,omitempty"`
	// Seed is a seed for deterministic sampling.
	Seed int `json:"seed,omitempty"`
}

// ResponseFormat specifies the format of the response.
type ResponseFormat struct {
	// Type is the format type (text, json_object).
	Type string `json:"type"`
}

// Tool represents a tool available to the model.
type Tool struct {
	// Type is the type of the tool (function).
	Type string `json:"type"`
	// Function is the function definition.
	Function *FunctionDefinition `json:"function,omitempty"`
}

// FunctionDefinition represents a function definition.
type FunctionDefinition struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is a description of the function.
	Description string `json:"description,omitempty"`
	// Parameters is the parameters of the function.
	Parameters any `json:"parameters,omitempty"`
}

// ToolChoice represents a tool choice.
type ToolChoice struct {
	// Type is the type of the tool (function).
	Type string `json:"type"`
	// Function is the function reference.
	Function *FunctionReference `json:"function,omitempty"`
}

// FunctionReference represents a function reference.
type FunctionReference struct {
	// Name is the name of the function.
	Name string `json:"name"`
}

// ChatResponse represents a response from the DeepSeek chat completions API.
type ChatResponse struct {
	// ID is the unique identifier for the chat completion.
	ID string `json:"id"`
	// Object is the object type (chat.completion).
	Object string `json:"object"`
	// Created is the Unix timestamp of when the chat completion was created.
	Created int64 `json:"created"`
	// Model is the model used for the chat completion.
	Model string `json:"model"`
	// SystemFingerprint is the fingerprint of the system configuration.
	SystemFingerprint string `json:"system_fingerprint"`
	// Choices is the list of chat completion choices.
	Choices []ChatResponseChoice `json:"choices"`
	// Usage is the token usage information.
	Usage *Usage `json:"usage,omitempty"`
}

// ChatResponseChoice represents a chat completion choice.
type ChatResponseChoice struct {
	// Index is the index of the choice.
	Index int `json:"index"`
	// Message is the chat completion message.
	Message ChatResponseMessage `json:"message"`
	// FinishReason is the reason the model stopped generating tokens.
	FinishReason string `json:"finish_reason"`
	// LogProbs is the log probabilities of the tokens.
	LogProbs *LogProbs `json:"logprobs,omitempty"`
}

// ChatResponseMessage represents a message in a chat completion response.
type ChatResponseMessage struct {
	// Role is the role of the message sender (assistant).
	Role string `json:"role"`
	// Content is the content of the message.
	Content string `json:"content"`
	// ReasoningContent is the reasoning content (for deepseek-reasoner model only).
	ReasoningContent string `json:"reasoning_content,omitempty"`
	// ToolCalls is the list of tool calls.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in a chat completion response.
type ToolCall struct {
	// ID is the unique identifier for the tool call.
	ID string `json:"id"`
	// Type is the type of the tool call (function).
	Type string `json:"type"`
	// Function is the function call.
	Function *FunctionCall `json:"function,omitempty"`
}

// FunctionCall represents a function call in a tool call.
type FunctionCall struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Arguments is the arguments to the function.
	Arguments string `json:"arguments"`
}

// LogProbs represents the log probabilities of tokens.
type LogProbs struct {
	// Content is the log probabilities of the content tokens.
	Content []TokenLogProb `json:"content,omitempty"`
}

// TokenLogProb represents the log probability of a token.
type TokenLogProb struct {
	// Token is the token string.
	Token string `json:"token"`
	// LogProb is the log probability of the token.
	LogProb float64 `json:"logprob"`
	// Bytes is the bytes of the token.
	Bytes []int `json:"bytes,omitempty"`
	// TopLogProbs is the list of top log probabilities.
	TopLogProbs []TopLogProb `json:"top_logprobs,omitempty"`
}

// TopLogProb represents a top log probability.
type TopLogProb struct {
	// Token is the token string.
	Token string `json:"token"`
	// LogProb is the log probability of the token.
	LogProb float64 `json:"logprob"`
	// Bytes is the bytes of the token.
	Bytes []int `json:"bytes,omitempty"`
}

// Usage represents the token usage information.
type Usage struct {
	// PromptTokens is the number of tokens in the prompt.
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is the number of tokens in the completion.
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is the total number of tokens used.
	TotalTokens int `json:"total_tokens"`
	// PromptCacheHit indicates whether the prompt was cached.
	PromptCacheHit bool `json:"prompt_cache_hit,omitempty"`
}

// StreamResponse represents a streaming response from the DeepSeek chat completions API.
type StreamResponse struct {
	// ID is the unique identifier for the chat completion.
	ID string `json:"id"`
	// Object is the object type (chat.completion.chunk).
	Object string `json:"object"`
	// Created is the Unix timestamp of when the chat completion was created.
	Created int64 `json:"created"`
	// Model is the model used for the chat completion.
	Model string `json:"model"`
	// SystemFingerprint is the fingerprint of the system configuration.
	SystemFingerprint string `json:"system_fingerprint"`
	// Choices is the list of chat completion choices.
	Choices []StreamResponseChoice `json:"choices"`
	// Usage is the token usage information.
	Usage *Usage `json:"usage,omitempty"`
}

// StreamResponseChoice represents a chat completion choice in a streaming response.
type StreamResponseChoice struct {
	// Index is the index of the choice.
	Index int `json:"index"`
	// Delta is the delta of the chat completion message.
	Delta StreamResponseDelta `json:"delta"`
	// FinishReason is the reason the model stopped generating tokens.
	FinishReason string `json:"finish_reason,omitempty"`
}

// StreamResponseDelta represents a delta of a chat completion message in a streaming response.
type StreamResponseDelta struct {
	// Role is the role of the message sender (assistant).
	Role string `json:"role,omitempty"`
	// Content is the content of the message.
	Content string `json:"content,omitempty"`
	// ReasoningContent is the reasoning content (for deepseek-reasoner model only).
	ReasoningContent string `json:"reasoning_content,omitempty"`
	// ToolCalls is the list of tool calls.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ErrorResponse represents an error response from the DeepSeek API.
type ErrorResponse struct {
	// Error is the error information.
	Error struct {
		// Message is the error message.
		Message string `json:"message"`
		// Type is the error type.
		Type string `json:"type"`
		// Param is the parameter that caused the error.
		Param string `json:"param,omitempty"`
		// Code is the error code.
		Code string `json:"code,omitempty"`
	} `json:"error"`
}
