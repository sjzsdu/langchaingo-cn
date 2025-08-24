package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Config 是整个应用的配置结构
type Config struct {
	LLMs       map[string]*LLMConfig       `json:"llms,omitempty"`
	Memories   map[string]*MemoryConfig    `json:"memories,omitempty"`
	Prompts    map[string]*PromptConfig    `json:"prompts,omitempty"`
	Embeddings map[string]*EmbeddingConfig `json:"embeddings,omitempty"`
	Chains     map[string]*ChainConfig     `json:"chains,omitempty"`
	Agents     map[string]*AgentConfig     `json:"agents,omitempty"`
}

// LLMConfig LLM组件配置
type LLMConfig struct {
	Type        string                 `json:"type"`        // openai, deepseek, kimi, qwen, anthropic, ollama
	Model       string                 `json:"model"`       // 模型名称
	APIKey      string                 `json:"api_key"`     // API密钥，支持环境变量
	BaseURL     string                 `json:"base_url"`    // 基础URL
	Temperature *float64               `json:"temperature"` // 温度参数
	MaxTokens   *int                   `json:"max_tokens"`  // 最大token数
	Options     map[string]interface{} `json:"options"`     // 其他选项
}

// MemoryConfig Memory组件配置
type MemoryConfig struct {
	Type           string                 `json:"type"`             // conversation_buffer, conversation_summary, conversation_token_buffer
	MaxTokenLimit  *int                   `json:"max_token_limit"`  // token限制
	MaxMessages    *int                   `json:"max_messages"`     // 消息数量限制
	ReturnMessages *bool                  `json:"return_messages"`  // 是否返回消息
	LLMRef         string                 `json:"llm_ref"`          // 引用的LLM组件
	Options        map[string]interface{} `json:"options"`          // 其他选项
}

// PromptConfig Prompt组件配置
type PromptConfig struct {
	Type              string                 `json:"type"`                // prompt_template, chat_prompt_template
	Template          string                 `json:"template"`            // 模板内容
	InputVariables    []string               `json:"input_variables"`     // 输入变量
	PartialVariables  map[string]string      `json:"partial_variables"`   // 部分变量
	TemplateFormat    string                 `json:"template_format"`     // 模板格式
	ValidateTemplate  *bool                  `json:"validate_template"`   // 是否验证模板
	Messages          []ChatMessageConfig    `json:"messages"`            // 聊天消息（用于chat_prompt_template）
	Options           map[string]interface{} `json:"options"`             // 其他选项
}

// ChatMessageConfig 聊天消息配置
type ChatMessageConfig struct {
	Role     string `json:"role"`     // system, human, ai
	Template string `json:"template"` // 消息模板
}

// EmbeddingConfig Embedding组件配置
type EmbeddingConfig struct {
	Type       string                 `json:"type"`        // openai, voyage, cohere
	Model      string                 `json:"model"`       // 模型名称
	APIKey     string                 `json:"api_key"`     // API密钥
	BaseURL    string                 `json:"base_url"`    // 基础URL
	BatchSize  *int                   `json:"batch_size"`  // 批处理大小
	Options    map[string]interface{} `json:"options"`     // 其他选项
}

// ChainConfig Chain组件配置
type ChainConfig struct {
	Type            string                 `json:"type"`             // llm, conversation, sequential, stuff_documents, map_reduce
	LLMRef          string                 `json:"llm_ref"`          // 引用的LLM组件
	MemoryRef       string                 `json:"memory_ref"`       // 引用的Memory组件
	PromptRef       string                 `json:"prompt_ref"`       // 引用的Prompt组件
	Chains          []string               `json:"chains"`           // 子链（用于sequential）
	InputKeys       []string               `json:"input_keys"`       // 输入键
	OutputKeys      []string               `json:"output_keys"`      // 输出键
	Separator       string                 `json:"separator"`        // 分隔符（用于stuff_documents）
	MaxConcurrency  *int                   `json:"max_concurrency"`  // 最大并发数
	Options         map[string]interface{} `json:"options"`          // 其他选项
}

// AgentConfig Agent组件配置
type AgentConfig struct {
	Type      string                 `json:"type"`       // zero_shot_react, conversational_react
	LLMRef    string                 `json:"llm_ref"`    // 引用的LLM组件
	MemoryRef string                 `json:"memory_ref"` // 引用的Memory组件
	Tools     []string               `json:"tools"`      // 工具列表
	MaxSteps  *int                   `json:"max_steps"`  // 最大步数
	Options   map[string]interface{} `json:"options"`    // 其他选项
}

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 替换环境变量
	data = []byte(expandEnvVars(string(data)))

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadConfigFromJSON 从JSON字符串加载配置
func LoadConfigFromJSON(jsonStr string) (*Config, error) {
	// 替换环境变量
	jsonStr = expandEnvVars(jsonStr)

	var config Config
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// expandEnvVars 展开环境变量，支持 ${VAR_NAME} 格式
func expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		return os.Getenv(key)
	})
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证LLM配置
	for name, llmConfig := range c.LLMs {
		if err := llmConfig.Validate(); err != nil {
			return fmt.Errorf("invalid LLM config '%s': %w", name, err)
		}
	}

	// 验证Memory配置
	for name, memoryConfig := range c.Memories {
		if err := memoryConfig.Validate(); err != nil {
			return fmt.Errorf("invalid Memory config '%s': %w", name, err)
		}
	}

	// 验证Prompt配置
	for name, promptConfig := range c.Prompts {
		if err := promptConfig.Validate(); err != nil {
			return fmt.Errorf("invalid Prompt config '%s': %w", name, err)
		}
	}

	// 验证Embedding配置
	for name, embeddingConfig := range c.Embeddings {
		if err := embeddingConfig.Validate(); err != nil {
			return fmt.Errorf("invalid Embedding config '%s': %w", name, err)
		}
	}

	// 验证Chain配置
	for name, chainConfig := range c.Chains {
		if err := chainConfig.ValidateReferences(c); err != nil {
			return fmt.Errorf("invalid Chain config '%s': %w", name, err)
		}
	}

	// 验证Agent配置
	for name, agentConfig := range c.Agents {
		if err := agentConfig.ValidateReferences(c); err != nil {
			return fmt.Errorf("invalid Agent config '%s': %w", name, err)
		}
	}

	return nil
}

// Validate 验证LLM配置
func (l *LLMConfig) Validate() error {
	if l.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"openai", "deepseek", "kimi", "qwen", "anthropic", "ollama"}
	if !contains(supportedTypes, l.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", l.Type, strings.Join(supportedTypes, ", "))
	}

	if l.Model == "" {
		return fmt.Errorf("model is required")
	}

	return nil
}

// Validate 验证Memory配置
func (m *MemoryConfig) Validate() error {
	if m.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"conversation_buffer", "conversation_summary", "conversation_token_buffer", "simple"}
	if !contains(supportedTypes, m.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", m.Type, strings.Join(supportedTypes, ", "))
	}

	return nil
}

// Validate 验证Prompt配置
func (p *PromptConfig) Validate() error {
	if p.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"prompt_template", "chat_prompt_template"}
	if !contains(supportedTypes, p.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", p.Type, strings.Join(supportedTypes, ", "))
	}

	if p.Type == "prompt_template" && p.Template == "" {
		return fmt.Errorf("template is required for prompt_template")
	}

	if p.Type == "chat_prompt_template" && len(p.Messages) == 0 {
		return fmt.Errorf("messages are required for chat_prompt_template")
	}

	return nil
}

// Validate 验证Embedding配置
func (e *EmbeddingConfig) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"openai", "voyage", "cohere"}
	if !contains(supportedTypes, e.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", e.Type, strings.Join(supportedTypes, ", "))
	}

	if e.Model == "" {
		return fmt.Errorf("model is required")
	}

	return nil
}

// ValidateReferences 验证Chain配置的引用
func (c *ChainConfig) ValidateReferences(config *Config) error {
	if c.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"llm", "conversation", "sequential", "stuff_documents", "map_reduce"}
	if !contains(supportedTypes, c.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", c.Type, strings.Join(supportedTypes, ", "))
	}

	// 验证LLM引用
	if c.LLMRef != "" {
		if _, exists := config.LLMs[c.LLMRef]; !exists {
			return fmt.Errorf("referenced LLM '%s' not found", c.LLMRef)
		}
	}

	// 验证Memory引用
	if c.MemoryRef != "" {
		if _, exists := config.Memories[c.MemoryRef]; !exists {
			return fmt.Errorf("referenced Memory '%s' not found", c.MemoryRef)
		}
	}

	// 验证Prompt引用
	if c.PromptRef != "" {
		if _, exists := config.Prompts[c.PromptRef]; !exists {
			return fmt.Errorf("referenced Prompt '%s' not found", c.PromptRef)
		}
	}

	return nil
}

// ValidateReferences 验证Agent配置的引用
func (a *AgentConfig) ValidateReferences(config *Config) error {
	if a.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"zero_shot_react", "conversational_react"}
	if !contains(supportedTypes, a.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", a.Type, strings.Join(supportedTypes, ", "))
	}

	// 验证LLM引用
	if a.LLMRef != "" {
		if _, exists := config.LLMs[a.LLMRef]; !exists {
			return fmt.Errorf("referenced LLM '%s' not found", a.LLMRef)
		}
	}

	// 验证Memory引用
	if a.MemoryRef != "" {
		if _, exists := config.Memories[a.MemoryRef]; !exists {
			return fmt.Errorf("referenced Memory '%s' not found", a.MemoryRef)
		}
	}

	return nil
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}