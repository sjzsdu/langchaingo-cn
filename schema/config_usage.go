package schema

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/agents"
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

// AgentUsageConfig Agent使用配置 - 简化版，对应实际Agent结构
type AgentUsageConfig struct {
	Type      string                 `json:"type"`                 // zero_shot_react, conversational_react
	Chain     *ChainUsageConfig      `json:"chain"`                // 核心处理链配置
	OutputKey string                 `json:"output_key,omitempty"` // 输出键，默认为"output"
	Options   map[string]interface{} `json:"options,omitempty"`    // 其他选项（如Tools、Callbacks等通过Options传递）
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

	if a.Chain == nil {
		return fmt.Errorf("chain is required")
	}

	if err := a.Chain.Validate(); err != nil {
		return fmt.Errorf("invalid chain config: %w", err)
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

// toChainConfig 将ChainUsageConfig转换为ChainConfig
func (c *ChainUsageConfig) toChainConfig(config *Config) (*ChainConfig, error) {
	chainConfig := &ChainConfig{
		Type:           c.Type,
		InputKeys:      c.InputKeys,
		OutputKeys:     c.OutputKeys,
		Separator:      c.Separator,
		MaxConcurrency: c.MaxConcurrency,
		Options:        c.Options,
	}

	// 处理LLM引用
	if c.LLM != nil {
		llmName := fmt.Sprintf("chain_llm_%p", c) // 使用指针地址确保唯一性
		config.LLMs[llmName] = c.LLM
		chainConfig.LLMRef = llmName
	}

	// 处理Memory引用
	if c.Memory != nil {
		memoryName := fmt.Sprintf("chain_memory_%p", c)
		memoryConfig := &MemoryConfig{
			Type:           c.Memory.Type,
			MaxTokenLimit:  c.Memory.MaxTokenLimit,
			MaxMessages:    c.Memory.MaxMessages,
			ReturnMessages: c.Memory.ReturnMessages,
			Options:        c.Memory.Options,
		}

		if c.Memory.LLM != nil {
			memoryLLMName := fmt.Sprintf("chain_memory_llm_%p", c)
			config.LLMs[memoryLLMName] = c.Memory.LLM
			memoryConfig.LLMRef = memoryLLMName
		}

		config.Memories[memoryName] = memoryConfig
		chainConfig.MemoryRef = memoryName
	}

	// 处理Prompt引用
	if c.Prompt != nil {
		promptName := fmt.Sprintf("chain_prompt_%p", c)
		config.Prompts[promptName] = c.Prompt
		chainConfig.PromptRef = promptName
	}

	// 处理子链
	if len(c.Chains) > 0 {
		subChainRefs := make([]string, len(c.Chains))
		for i, subChain := range c.Chains {
			subChainName := fmt.Sprintf("subchain_%d_%p", i, c)
			subChainConfig, err := subChain.toChainConfig(config)
			if err != nil {
				return nil, fmt.Errorf("failed to convert subchain[%d]: %w", i, err)
			}
			config.Chains[subChainName] = subChainConfig
			subChainRefs[i] = subChainName
		}
		chainConfig.Chains = subChainRefs
	}

	return chainConfig, nil
}

// ToConfig 将ExecutorUsageConfig转换为原始的Config格式
func (e *ExecutorUsageConfig) ToConfig() (*Config, error) {
	config := &Config{
		LLMs:      make(map[string]*LLMConfig),
		Memories:  make(map[string]*MemoryConfig),
		Prompts:   make(map[string]*PromptConfig),
		Chains:    make(map[string]*ChainConfig),
		Agents:    make(map[string]*AgentConfig),
		Executors: make(map[string]*ExecutorConfig),
	}

	if e.Agent != nil {
		// 创建Agent配置
		agentName := "main_agent"
		agentConfig := &AgentConfig{
			Type:    e.Agent.Type,
			Options: make(map[string]interface{}),
		}

		// 复制原有的Options
		if e.Agent.Options != nil {
			for k, v := range e.Agent.Options {
				agentConfig.Options[k] = v
			}
		}

		// 处理Agent的Chain配置
		if e.Agent.Chain != nil {
			// 创建完整的chain配置
			chainName := "agent_chain"
			chainConfig, err := e.Agent.Chain.toChainConfig(config)
			if err != nil {
				return nil, fmt.Errorf("failed to convert agent chain config: %w", err)
			}
			config.Chains[chainName] = chainConfig
			agentConfig.ChainRef = chainName
		}

		// 设置OutputKey
		if e.Agent.OutputKey != "" {
			agentConfig.OutputKey = e.Agent.OutputKey
		}

		config.Agents[agentName] = agentConfig

		// 处理Executor级别的Memory
		var executorMemoryName string
		if e.Memory != nil {
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
		}

		config.Executors["main_executor"] = executorConfig
	}

	return config, nil
}

// CreateExecutor 从ExecutorUsageConfig直接创建Executor实例
func (e *ExecutorUsageConfig) CreateExecutor() (*agents.Executor, error) {
	// Validate config first
	if err := e.Validate(); err != nil {
		return nil, fmt.Errorf("invalid executor config: %w", err)
	}

	// Create factory
	factory := NewFactory()

	// Convert to standard config format
	config, err := e.ToConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}

	// Get the main agent config referenced by executor
	executorConfig, exists := config.Executors["main_executor"]
	if !exists {
		return nil, fmt.Errorf("main_executor not found in config")
	}

	agentConfig, exists := config.Agents[executorConfig.AgentRef]
	if !exists {
		return nil, fmt.Errorf("agent '%s' not found in config", executorConfig.AgentRef)
	}

	// Create executor using factory
	executor, err := factory.CreateAgent(agentConfig, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	return executor, nil
}

// CreateExecutorFromFile 从文件创建Executor实例
func CreateExecutorFromFile(filename string) (*agents.Executor, error) {
	config, err := LoadExecutorUsageConfigFromFile(filename)
	if err != nil {
		return nil, err
	}

	return config.CreateExecutor()
}

// CreateExecutorFromJSON 从JSON字符串创建Executor实例
func CreateExecutorFromJSON(jsonStr string) (*agents.Executor, error) {
	config, err := LoadExecutorUsageConfigFromJSON(jsonStr)
	if err != nil {
		return nil, err
	}

	return config.CreateExecutor()
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
