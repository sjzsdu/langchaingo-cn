package schema

import (
	"encoding/json"
	"fmt"
	"os"
)

// ExecutorUsageConfig 以Executor为入口的使用风格配置
// 这种风格不使用引用(ref)，而是直接嵌入完整的配置对象
type ExecutorUsageConfig struct {
	Agent                   *AgentUsageConfig      `json:"agent"`                               // 直接嵌入Agent配置
	Memory                  *MemoryUsageConfig     `json:"memory,omitempty"`                    // 直接嵌入Memory配置
	MaxIterations           *int                   `json:"max_iterations,omitempty"`            // 最大迭代次数
	ReturnIntermediateSteps *bool                  `json:"return_intermediate_steps,omitempty"` // 是否返回中间步骤
	ErrorHandler            *ErrorHandlerConfig    `json:"error_handler,omitempty"`             // 错误处理器配置
	Options                 map[string]interface{} `json:"options,omitempty"`                   // 其他选项
}

// AgentUsageConfig Agent使用配置 - 直接嵌入依赖
type AgentUsageConfig struct {
	Type     string                 `json:"type"`                // zero_shot_react, conversational_react
	LLM      *LLMConfig             `json:"llm"`                 // 直接嵌入LLM配置
	Memory   *MemoryUsageConfig     `json:"memory,omitempty"`    // 直接嵌入Memory配置
	MaxSteps *int                   `json:"max_steps,omitempty"` // 最大步数
	Options  map[string]interface{} `json:"options,omitempty"`   // 其他选项
}

// MemoryUsageConfig Memory使用配置 - 直接嵌入依赖
type MemoryUsageConfig struct {
	Type           string                 `json:"type"`                      // conversation_buffer, conversation_summary, conversation_token_buffer
	MaxTokenLimit  *int                   `json:"max_token_limit,omitempty"` // token限制
	MaxMessages    *int                   `json:"max_messages,omitempty"`    // 消息数量限制
	ReturnMessages *bool                  `json:"return_messages,omitempty"` // 是否返回消息
	LLM            *LLMConfig             `json:"llm,omitempty"`             // 直接嵌入LLM配置（用于summary类型）
	Options        map[string]interface{} `json:"options,omitempty"`         // 其他选项
}

// ChainUsageConfig Chain使用配置 - 直接嵌入依赖
type ChainUsageConfig struct {
	Type           string                 `json:"type"`                      // llm, conversation, sequential, stuff_documents, map_reduce
	LLM            *LLMConfig             `json:"llm,omitempty"`             // 直接嵌入LLM配置
	Memory         *MemoryUsageConfig     `json:"memory,omitempty"`          // 直接嵌入Memory配置
	Prompt         *PromptConfig          `json:"prompt,omitempty"`          // 直接嵌入Prompt配置
	Chains         []*ChainUsageConfig    `json:"chains,omitempty"`          // 子链列表（用于sequential）
	InputKeys      []string               `json:"input_keys,omitempty"`      // 输入键
	OutputKeys     []string               `json:"output_keys,omitempty"`     // 输出键
	Separator      string                 `json:"separator,omitempty"`       // 分隔符（用于stuff_documents）
	MaxConcurrency *int                   `json:"max_concurrency,omitempty"` // 最大并发数
	Options        map[string]interface{} `json:"options,omitempty"`         // 其他选项
}

// EmbeddingUsageConfig Embedding使用配置
type EmbeddingUsageConfig struct {
	Type      string                 `json:"type"`                 // openai, voyage, cohere
	Model     string                 `json:"model"`                // 模型名称
	APIKey    string                 `json:"api_key"`              // API密钥
	BaseURL   string                 `json:"base_url,omitempty"`   // 基础URL
	BatchSize *int                   `json:"batch_size,omitempty"` // 批处理大小
	Options   map[string]interface{} `json:"options,omitempty"`    // 其他选项
}

// LoadExecutorUsageConfigFromFile 从文件加载Executor使用配置
func LoadExecutorUsageConfigFromFile(filename string) (*ExecutorUsageConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 替换环境变量
	data = []byte(expandEnvVars(string(data)))

	var config ExecutorUsageConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadExecutorUsageConfigFromJSON 从JSON字符串加载Executor使用配置
func LoadExecutorUsageConfigFromJSON(jsonStr string) (*ExecutorUsageConfig, error) {
	// 替换环境变量
	jsonStr = expandEnvVars(jsonStr)

	var config ExecutorUsageConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// LoadChainUsageConfigFromFile 从文件加载Chain使用配置
func LoadChainUsageConfigFromFile(filename string) (*ChainUsageConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 替换环境变量
	data = []byte(expandEnvVars(string(data)))

	var config ChainUsageConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// LoadChainUsageConfigFromJSON 从JSON字符串加载Chain使用配置
func LoadChainUsageConfigFromJSON(jsonStr string) (*ChainUsageConfig, error) {
	// 替换环境变量
	jsonStr = expandEnvVars(jsonStr)

	var config ChainUsageConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// Validate 验证ExecutorUsageConfig的有效性
func (e *ExecutorUsageConfig) Validate() error {
	if e.Agent == nil {
		return fmt.Errorf("agent is required")
	}

	if err := e.Agent.Validate(); err != nil {
		return fmt.Errorf("invalid agent config: %w", err)
	}

	if e.Memory != nil {
		if err := e.Memory.Validate(); err != nil {
			return fmt.Errorf("invalid memory config: %w", err)
		}
	}

	if e.ErrorHandler != nil {
		if err := e.ErrorHandler.Validate(); err != nil {
			return fmt.Errorf("invalid error handler config: %w", err)
		}
	}

	return nil
}

// Validate 验证AgentUsageConfig的有效性
func (a *AgentUsageConfig) Validate() error {
	if a.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"zero_shot_react", "conversational_react"}
	if !contains(supportedTypes, a.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", a.Type, joinStrings(supportedTypes, ", "))
	}

	if a.LLM == nil {
		return fmt.Errorf("llm is required")
	}

	if err := a.LLM.Validate(); err != nil {
		return fmt.Errorf("invalid llm config: %w", err)
	}

	if a.Memory != nil {
		if err := a.Memory.Validate(); err != nil {
			return fmt.Errorf("invalid memory config: %w", err)
		}
	}

	return nil
}

// Validate 验证MemoryUsageConfig的有效性
func (m *MemoryUsageConfig) Validate() error {
	if m.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"conversation_buffer", "conversation_summary", "conversation_token_buffer", "simple"}
	if !contains(supportedTypes, m.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", m.Type, joinStrings(supportedTypes, ", "))
	}

	// 对于summary类型的memory，需要LLM
	if m.Type == "conversation_summary" && m.LLM == nil {
		return fmt.Errorf("llm is required for conversation_summary memory type")
	}

	if m.LLM != nil {
		if err := m.LLM.Validate(); err != nil {
			return fmt.Errorf("invalid llm config: %w", err)
		}
	}

	return nil
}

// Validate 验证ChainUsageConfig的有效性
func (c *ChainUsageConfig) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"llm", "conversation", "sequential", "stuff_documents", "map_reduce"}
	if !contains(supportedTypes, c.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", c.Type, joinStrings(supportedTypes, ", "))
	}

	// 根据不同类型验证必需的依赖
	switch c.Type {
	case "llm":
		if c.LLM == nil {
			return fmt.Errorf("llm is required for llm chain type")
		}
		if err := c.LLM.Validate(); err != nil {
			return fmt.Errorf("invalid llm config: %w", err)
		}
	case "conversation":
		if c.LLM == nil {
			return fmt.Errorf("llm is required for conversation chain type")
		}
		if err := c.LLM.Validate(); err != nil {
			return fmt.Errorf("invalid llm config: %w", err)
		}
		if c.Memory != nil {
			if err := c.Memory.Validate(); err != nil {
				return fmt.Errorf("invalid memory config: %w", err)
			}
		}
	case "sequential":
		if len(c.Chains) == 0 {
			return fmt.Errorf("chains are required for sequential chain type")
		}
		for i, chain := range c.Chains {
			if err := chain.Validate(); err != nil {
				return fmt.Errorf("invalid chain[%d] config: %w", i, err)
			}
		}
	}

	if c.Prompt != nil {
		if err := c.Prompt.Validate(); err != nil {
			return fmt.Errorf("invalid prompt config: %w", err)
		}
	}

	return nil
}

// Validate 验证EmbeddingUsageConfig的有效性
func (e *EmbeddingUsageConfig) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("type is required")
	}

	supportedTypes := []string{"openai", "voyage", "cohere"}
	if !contains(supportedTypes, e.Type) {
		return fmt.Errorf("unsupported type: %s, supported: %s", e.Type, joinStrings(supportedTypes, ", "))
	}

	if e.Model == "" {
		return fmt.Errorf("model is required")
	}

	return nil
}

// ToConfig 将ExecutorUsageConfig转换为原始的Config格式
func (e *ExecutorUsageConfig) ToConfig() (*Config, error) {
	config := &Config{
		LLMs:      make(map[string]*LLMConfig),
		Memories:  make(map[string]*MemoryConfig),
		Prompts:   make(map[string]*PromptConfig),
		Agents:    make(map[string]*AgentConfig),
		Executors: make(map[string]*ExecutorConfig),
	}

	// 提取并注册Agent的LLM
	if e.Agent != nil && e.Agent.LLM != nil {
		llmName := "agent_llm"
		config.LLMs[llmName] = e.Agent.LLM

		// 提取并注册Agent的Memory
		var memoryName string
		if e.Agent.Memory != nil {
			memoryName = "agent_memory"
			memoryConfig := &MemoryConfig{
				Type:           e.Agent.Memory.Type,
				MaxTokenLimit:  e.Agent.Memory.MaxTokenLimit,
				MaxMessages:    e.Agent.Memory.MaxMessages,
				ReturnMessages: e.Agent.Memory.ReturnMessages,
				Options:        e.Agent.Memory.Options,
			}

			// 如果Memory需要LLM，设置引用
			if e.Agent.Memory.LLM != nil {
				memoryLLMName := "memory_llm"
				config.LLMs[memoryLLMName] = e.Agent.Memory.LLM
				memoryConfig.LLMRef = memoryLLMName
			}

			config.Memories[memoryName] = memoryConfig
		}

		// 创建Agent配置
		agentName := "main_agent"
		agentConfig := &AgentConfig{
			Type:     e.Agent.Type,
			LLMRef:   llmName,
			MaxSteps: e.Agent.MaxSteps,
			Options:  e.Agent.Options,
		}
		if memoryName != "" {
			agentConfig.MemoryRef = memoryName
		}
		config.Agents[agentName] = agentConfig

		// 如果Executor有自己的Memory
		var executorMemoryName string
		if e.Memory != nil && e.Memory != e.Agent.Memory {
			executorMemoryName = "executor_memory"
			executorMemoryConfig := &MemoryConfig{
				Type:           e.Memory.Type,
				MaxTokenLimit:  e.Memory.MaxTokenLimit,
				MaxMessages:    e.Memory.MaxMessages,
				ReturnMessages: e.Memory.ReturnMessages,
				Options:        e.Memory.Options,
			}

			if e.Memory.LLM != nil {
				executorMemoryLLMName := "executor_memory_llm"
				config.LLMs[executorMemoryLLMName] = e.Memory.LLM
				executorMemoryConfig.LLMRef = executorMemoryLLMName
			}

			config.Memories[executorMemoryName] = executorMemoryConfig
		}

		// 创建Executor配置
		executorConfig := &ExecutorConfig{
			AgentRef:                agentName,
			MaxIterations:           e.MaxIterations,
			ReturnIntermediateSteps: e.ReturnIntermediateSteps,
			ErrorHandlerConfig:      e.ErrorHandler,
			Options:                 e.Options,
		}
		if executorMemoryName != "" {
			executorConfig.MemoryRef = executorMemoryName
		} else if memoryName != "" {
			executorConfig.MemoryRef = memoryName
		}

		config.Executors["main_executor"] = executorConfig
	}

	return config, nil
}

// joinStrings 连接字符串切片
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
