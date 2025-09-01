package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigGenerator 配置生成器，提供快速生成配置文件的门面方法
type ConfigGenerator struct {
	outputDir string // 输出目录
}

// NewConfigGenerator 创建配置生成器
func NewConfigGenerator(outputDir string) *ConfigGenerator {
	if outputDir == "" {
		outputDir = "."
	}
	return &ConfigGenerator{
		outputDir: outputDir,
	}
}

// LLMTemplate LLM配置模板参数
type LLMTemplate struct {
	Type        string  // 必需：LLM类型 (openai, deepseek, kimi, qwen, anthropic, ollama)
	Model       string  // 必需：模型名称
	APIKey      string  // API密钥（默认使用环境变量）
	BaseURL     string  // 可选：自定义API基础URL
	Temperature float64 // 可选：温度参数 (0.0-2.0)
	MaxTokens   int     // 可选：最大token数
}

// ChainTemplate Chain配置模板参数
type ChainTemplate struct {
	Type           string      // 必需：Chain类型 (llm, conversation, sequential)
	LLMTemplate    LLMTemplate // LLM配置
	MemoryType     string      // 可选：Memory类型 (conversation_buffer, simple)
	PromptTemplate string      // 可选：提示模板
	InputVariables []string    // 可选：输入变量
}

// AgentTemplate Agent配置模板参数
type AgentTemplate struct {
	Type        string      // 必需：Agent类型 (zero_shot_react, conversational_react)
	LLMTemplate LLMTemplate // LLM配置
	MemoryType  string      // 可选：Memory类型 (默认 conversation_buffer)
	MaxSteps    int         // 可选：最大步数 (默认 5)
}

// ExecutorTemplate Executor配置模板参数
type ExecutorTemplate struct {
	AgentTemplate           AgentTemplate // Agent配置
	MaxIterations           int           // 可选：最大迭代次数 (默认 10)
	ReturnIntermediateSteps bool          // 可选：是否返回中间步骤
}

// GenerateLLMConfig 生成极简LLM配置文件
func (g *ConfigGenerator) GenerateLLMConfig(template LLMTemplate, filename string) error {
	if template.Type == "" {
		return fmt.Errorf("LLM type is required")
	}
	if template.Model == "" {
		return fmt.Errorf("model is required")
	}

	// 设置默认API密钥环境变量
	if template.APIKey == "" {
		template.APIKey = g.getDefaultAPIKeyEnv(template.Type)
	}

	// 创建配置
	config := &Config{
		LLMs: map[string]*LLMConfig{
			"main_llm": {
				Type:   template.Type,
				Model:  template.Model,
				APIKey: template.APIKey,
			},
		},
	}

	// 设置可选参数
	if template.BaseURL != "" {
		config.LLMs["main_llm"].BaseURL = template.BaseURL
	}
	if template.Temperature > 0 {
		config.LLMs["main_llm"].Temperature = &template.Temperature
	}
	if template.MaxTokens > 0 {
		config.LLMs["main_llm"].MaxTokens = &template.MaxTokens
	}

	return g.writeConfigToFile(config, filename)
}

// GenerateChainConfig 生成极简Chain配置文件
func (g *ConfigGenerator) GenerateChainConfig(template ChainTemplate, filename string) error {
	if template.Type == "" {
		return fmt.Errorf("Chain type is required")
	}
	if template.LLMTemplate.Type == "" {
		return fmt.Errorf("LLM type is required")
	}
	if template.LLMTemplate.Model == "" {
		return fmt.Errorf("LLM model is required")
	}

	// 设置默认API密钥
	if template.LLMTemplate.APIKey == "" {
		template.LLMTemplate.APIKey = g.getDefaultAPIKeyEnv(template.LLMTemplate.Type)
	}

	// 创建配置
	config := &Config{
		LLMs: map[string]*LLMConfig{
			"chain_llm": {
				Type:   template.LLMTemplate.Type,
				Model:  template.LLMTemplate.Model,
				APIKey: template.LLMTemplate.APIKey,
			},
		},
		Chains: map[string]*ChainConfig{
			"main_chain": {
				Type:   template.Type,
				LLMRef: "chain_llm",
			},
		},
	}

	// 设置LLM可选参数
	if template.LLMTemplate.BaseURL != "" {
		config.LLMs["chain_llm"].BaseURL = template.LLMTemplate.BaseURL
	}
	if template.LLMTemplate.Temperature > 0 {
		config.LLMs["chain_llm"].Temperature = &template.LLMTemplate.Temperature
	}
	if template.LLMTemplate.MaxTokens > 0 {
		config.LLMs["chain_llm"].MaxTokens = &template.LLMTemplate.MaxTokens
	}

	// 添加Memory（如果需要）
	if template.MemoryType != "" {
		if config.Memories == nil {
			config.Memories = make(map[string]*MemoryConfig)
		}
		config.Memories["chain_memory"] = &MemoryConfig{
			Type: template.MemoryType,
		}
		config.Chains["main_chain"].MemoryRef = "chain_memory"
	}

	// 添加Prompt（如果提供）
	if template.PromptTemplate != "" {
		if config.Prompts == nil {
			config.Prompts = make(map[string]*PromptConfig)
		}
		inputVars := template.InputVariables
		if inputVars == nil {
			inputVars = []string{"input"}
		}
		config.Prompts["chain_prompt"] = &PromptConfig{
			Type:           "prompt_template",
			Template:       template.PromptTemplate,
			InputVariables: inputVars,
		}
		config.Chains["main_chain"].PromptRef = "chain_prompt"
	}

	return g.writeConfigToFile(config, filename)
}

// GenerateAgentConfig 生成极简Agent配置文件
func (g *ConfigGenerator) GenerateAgentConfig(template AgentTemplate, filename string) error {
	if template.Type == "" {
		return fmt.Errorf("Agent type is required")
	}
	if template.LLMTemplate.Type == "" {
		return fmt.Errorf("LLM type is required")
	}
	if template.LLMTemplate.Model == "" {
		return fmt.Errorf("LLM model is required")
	}

	// 设置默认值
	if template.LLMTemplate.APIKey == "" {
		template.LLMTemplate.APIKey = g.getDefaultAPIKeyEnv(template.LLMTemplate.Type)
	}
	if template.MemoryType == "" {
		template.MemoryType = "conversation_buffer"
	}
	if template.MaxSteps == 0 {
		template.MaxSteps = 5
	}

	// 创建配置
	config := &Config{
		LLMs: map[string]*LLMConfig{
			"agent_llm": {
				Type:   template.LLMTemplate.Type,
				Model:  template.LLMTemplate.Model,
				APIKey: template.LLMTemplate.APIKey,
			},
		},
		Chains: map[string]*ChainConfig{
			"agent_chain": {
				Type:   "conversation",
				LLMRef: "agent_llm",
			},
		},
		Agents: map[string]*AgentConfig{
			"main_agent": {
				Type:     template.Type,
				ChainRef: "agent_chain",
			},
		},
	}

	// 对于conversation类型的agent，添加memory配置
	if template.Type == "conversational_react" {
		config.Memories = map[string]*MemoryConfig{
			"agent_memory": {
				Type: template.MemoryType,
			},
		}
		config.Chains["agent_chain"].MemoryRef = "agent_memory"
	} else {
		// 对于零样本agent，使用简单的llm chain
		config.Chains["agent_chain"].Type = "llm"
	}

	// 设置LLM可选参数
	if template.LLMTemplate.BaseURL != "" {
		config.LLMs["agent_llm"].BaseURL = template.LLMTemplate.BaseURL
	}
	if template.LLMTemplate.Temperature > 0 {
		config.LLMs["agent_llm"].Temperature = &template.LLMTemplate.Temperature
	}
	if template.LLMTemplate.MaxTokens > 0 {
		config.LLMs["agent_llm"].MaxTokens = &template.LLMTemplate.MaxTokens
	}

	return g.writeConfigToFile(config, filename)
}

// GenerateExecutorConfig 生成极简Executor配置文件（使用新的usage风格）
func (g *ConfigGenerator) GenerateExecutorConfig(template ExecutorTemplate, filename string) error {
	if template.AgentTemplate.Type == "" {
		return fmt.Errorf("Agent type is required")
	}
	if template.AgentTemplate.LLMTemplate.Type == "" {
		return fmt.Errorf("LLM type is required")
	}
	if template.AgentTemplate.LLMTemplate.Model == "" {
		return fmt.Errorf("LLM model is required")
	}

	// 设置默认值
	if template.AgentTemplate.LLMTemplate.APIKey == "" {
		template.AgentTemplate.LLMTemplate.APIKey = g.getDefaultAPIKeyEnv(template.AgentTemplate.LLMTemplate.Type)
	}
	if template.AgentTemplate.MemoryType == "" {
		template.AgentTemplate.MemoryType = "conversation_buffer"
	}
	if template.AgentTemplate.MaxSteps == 0 {
		template.AgentTemplate.MaxSteps = 5
	}
	if template.MaxIterations == 0 {
		template.MaxIterations = 10
	}

	// 创建使用新风格的配置
	config := &ExecutorUsageConfig{
		Agent: &AgentUsageConfig{
			Type: template.AgentTemplate.Type,
			Chain: &ChainUsageConfig{
				Type: "llm", // Agent内部使用LLM Chain
				LLM: &LLMConfig{
					Type:   template.AgentTemplate.LLMTemplate.Type,
					Model:  template.AgentTemplate.LLMTemplate.Model,
					APIKey: template.AgentTemplate.LLMTemplate.APIKey,
				},
				Memory: &MemoryUsageConfig{
					Type: template.AgentTemplate.MemoryType,
				},
			},
			Options: map[string]interface{}{
				"max_steps": template.AgentTemplate.MaxSteps,
			},
		},
		MaxIterations:           &template.MaxIterations,
		ReturnIntermediateSteps: &template.ReturnIntermediateSteps,
	}

	// 设置LLM可选参数
	if template.AgentTemplate.LLMTemplate.BaseURL != "" {
		config.Agent.Chain.LLM.BaseURL = template.AgentTemplate.LLMTemplate.BaseURL
	}
	if template.AgentTemplate.LLMTemplate.Temperature > 0 {
		config.Agent.Chain.LLM.Temperature = &template.AgentTemplate.LLMTemplate.Temperature
	}
	if template.AgentTemplate.LLMTemplate.MaxTokens > 0 {
		config.Agent.Chain.LLM.MaxTokens = &template.AgentTemplate.LLMTemplate.MaxTokens
	}

	return g.writeExecutorConfigToFile(config, filename)
}

// 快捷方法：生成常用配置文件

// GenerateDeepSeekChatConfig 生成DeepSeek聊天配置
func (g *ConfigGenerator) GenerateDeepSeekChatConfig(filename string) error {
	return g.GenerateChainConfig(ChainTemplate{
		Type: "conversation",
		LLMTemplate: LLMTemplate{
			Type:        "deepseek",
			Model:       "deepseek-chat",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
		MemoryType: "conversation_buffer",
	}, filename)
}

// GenerateKimiChatConfig 生成Kimi聊天配置
func (g *ConfigGenerator) GenerateKimiChatConfig(filename string) error {
	return g.GenerateChainConfig(ChainTemplate{
		Type: "conversation",
		LLMTemplate: LLMTemplate{
			Type:        "kimi",
			Model:       "moonshot-v1-8k",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
		MemoryType: "conversation_buffer",
	}, filename)
}

// GenerateOpenAIChatConfig 生成OpenAI聊天配置
func (g *ConfigGenerator) GenerateOpenAIChatConfig(filename string) error {
	return g.GenerateChainConfig(ChainTemplate{
		Type: "conversation",
		LLMTemplate: LLMTemplate{
			Type:        "openai",
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
		MemoryType: "conversation_buffer",
	}, filename)
}

// GenerateQwenChatConfig 生成Qwen聊天配置
func (g *ConfigGenerator) GenerateQwenChatConfig(filename string) error {
	return g.GenerateChainConfig(ChainTemplate{
		Type: "conversation",
		LLMTemplate: LLMTemplate{
			Type:        "qwen",
			Model:       "qwen-plus",
			Temperature: 0.7,
			MaxTokens:   2048,
		},
		MemoryType: "conversation_buffer",
	}, filename)
}

// GenerateReactAgentConfig 生成零样本ReAct智能体配置
func (g *ConfigGenerator) GenerateReactAgentConfig(llmType, model, filename string) error {
	return g.GenerateAgentConfig(AgentTemplate{
		Type: "zero_shot_react",
		LLMTemplate: LLMTemplate{
			Type:        llmType,
			Model:       model,
			Temperature: 0.3,
			MaxTokens:   2048,
		},
		MemoryType: "conversation_buffer",
		MaxSteps:   5,
	}, filename)
}

// GenerateConversationalAgentConfig 生成对话智能体配置
func (g *ConfigGenerator) GenerateConversationalAgentConfig(llmType, model, filename string) error {
	return g.GenerateAgentConfig(AgentTemplate{
		Type: "conversational_react",
		LLMTemplate: LLMTemplate{
			Type:        llmType,
			Model:       model,
			Temperature: 0.3,
			MaxTokens:   2048,
		},
		MemoryType: "conversation_buffer",
		MaxSteps:   5,
	}, filename)
}

// GenerateExecutorWithDeepSeek 生成基于DeepSeek的执行器配置
func (g *ConfigGenerator) GenerateExecutorWithDeepSeek(filename string) error {
	return g.GenerateExecutorConfig(ExecutorTemplate{
		AgentTemplate: AgentTemplate{
			Type: "zero_shot_react",
			LLMTemplate: LLMTemplate{
				Type:        "deepseek",
				Model:       "deepseek-chat",
				Temperature: 0.7,
				MaxTokens:   2000,
			},
			MemoryType: "conversation_buffer",
			MaxSteps:   5,
		},
		MaxIterations:           10,
		ReturnIntermediateSteps: true,
	}, filename)
}

// 辅助方法

// getDefaultAPIKeyEnv 获取默认API密钥环境变量名
func (g *ConfigGenerator) getDefaultAPIKeyEnv(llmType string) string {
	switch llmType {
	case "openai":
		return "${OPENAI_API_KEY}"
	case "deepseek":
		return "${DEEPSEEK_API_KEY}"
	case "kimi":
		return "${KIMI_API_KEY}"
	case "qwen":
		return "${QWEN_API_KEY}"
	case "anthropic":
		return "${ANTHROPIC_API_KEY}"
	default:
		return ""
	}
}

// writeConfigToFile 将配置写入文件
func (g *ConfigGenerator) writeConfigToFile(config *Config, filename string) error {
	fullPath := filepath.Join(g.outputDir, filename)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 序列化为JSON（美化格式）
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ 配置文件已生成: %s\n", fullPath)
	return nil
}

// writeExecutorConfigToFile 将执行器配置写入文件
func (g *ConfigGenerator) writeExecutorConfigToFile(config *ExecutorUsageConfig, filename string) error {
	fullPath := filepath.Join(g.outputDir, filename)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 序列化为JSON（美化格式）
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ 执行器配置文件已生成: %s\n", fullPath)
	return nil
}

// 全局门面方法，使用默认生成器

var defaultGenerator = NewConfigGenerator(".")

// QuickGenerateLLM 快速生成LLM配置文件
func QuickGenerateLLM(llmType, model, filename string) error {
	return defaultGenerator.GenerateLLMConfig(LLMTemplate{
		Type:  llmType,
		Model: model,
	}, filename)
}

// QuickGenerateChain 快速生成Chain配置文件
func QuickGenerateChain(chainType, llmType, model, filename string) error {
	return defaultGenerator.GenerateChainConfig(ChainTemplate{
		Type: chainType,
		LLMTemplate: LLMTemplate{
			Type:  llmType,
			Model: model,
		},
	}, filename)
}

// QuickGenerateAgent 快速生成Agent配置文件
func QuickGenerateAgent(agentType, llmType, model, filename string) error {
	return defaultGenerator.GenerateAgentConfig(AgentTemplate{
		Type: agentType,
		LLMTemplate: LLMTemplate{
			Type:  llmType,
			Model: model,
		},
	}, filename)
}

// QuickGenerateExecutor 快速生成Executor配置文件
func QuickGenerateExecutor(agentType, llmType, model, filename string) error {
	return defaultGenerator.GenerateExecutorConfig(ExecutorTemplate{
		AgentTemplate: AgentTemplate{
			Type: agentType,
			LLMTemplate: LLMTemplate{
				Type:  llmType,
				Model: model,
			},
		},
	}, filename)
}
