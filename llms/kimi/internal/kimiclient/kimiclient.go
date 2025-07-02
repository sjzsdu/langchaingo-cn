package kimiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	// DefaultBaseURL 是Kimi API的默认基础URL
	DefaultBaseURL = "https://api.moonshot.cn/v1"

	// 默认模型
	defaultModel = "moonshot-v1-8k"
)

// ErrEmptyResponse 当Kimi API返回空响应时返回此错误
var ErrEmptyResponse = errors.New("空响应")

// Client 是Kimi API的客户端
type Client struct {
	token   string
	Model   string
	baseURL string

	httpClient Doer
}

// Option 是Kimi客户端的选项
type Option func(*Client) error

// Doer 执行HTTP请求
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// WithHTTPClient 允许设置自定义HTTP客户端
func WithHTTPClient(client Doer) Option {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

// WithBaseURL 设置API的基础URL
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
		return nil
	}
}

// New 返回一个新的Kimi客户端
func New(token string, model string, baseURL string, opts ...Option) (*Client, error) {
	c := &Client{
		Model:      model,
		token:      token,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// ChatRequest 是创建聊天请求的结构体
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Tools       []Tool        `json:"tools,omitempty"`
	ToolChoice  interface{}   `json:"tool_choice,omitempty"`

	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

// ChatMessage 是聊天消息
type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// Tool 是工具定义
type Tool struct {
	Type     string           `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionDefinition 是函数定义
type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

// ChatResponse 是聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatResponseChunk 是流式聊天响应的块
type ChatResponseChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Delta        map[string]interface{} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func (c *Client) setDefaults(payload *ChatRequest) {
	// 设置默认值
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 2048
	}

	switch {
	// 优先使用payload中指定的模型
	case payload.Model != "":

	// 如果payload中没有设置模型，使用客户端中指定的模型
	case c.Model != "":
		payload.Model = c.Model
	// 回退：使用默认模型
	default:
		payload.Model = defaultModel
	}

	if payload.StreamingFunc != nil {
		payload.Stream = true
	}
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
}

func (c *Client) do(ctx context.Context, path string, payloadBytes []byte) (*http.Response, error) {
	if c.baseURL == "" {
		c.baseURL = DefaultBaseURL
	}

	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	return resp, nil
}

type errorMessage struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (c *Client) decodeError(resp *http.Response) error {
	var errResp errorMessage
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("解析错误响应失败: %w", err)
	}

	return fmt.Errorf("API错误 (%d): %s - %s", resp.StatusCode, errResp.Error.Type, errResp.Error.Message)
}