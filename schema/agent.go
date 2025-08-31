package schema

import (
	"fmt"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

// AgentFactory Agent组件工厂
type AgentFactory struct {
	llmFactory    *LLMFactory
	memoryFactory *MemoryFactory
}

// NewAgentFactory 创建Agent工厂实例
func NewAgentFactory(llmFactory *LLMFactory, memoryFactory *MemoryFactory) *AgentFactory {
	return &AgentFactory{
		llmFactory:    llmFactory,
		memoryFactory: memoryFactory,
	}
}

// Create 根据配置创建Agent实例
func (f *AgentFactory) Create(config *AgentConfig, allConfigs *Config) (*agents.Executor, error) {
	if config == nil {
		return nil, fmt.Errorf("agent config is nil")
	}

	// 验证配置
	if err := config.ValidateReferences(allConfigs); err != nil {
		return nil, fmt.Errorf("invalid Agent config: %w", err)
	}

	switch config.Type {
	case "zero_shot_react":
		return f.createZeroShotReactAgent(config, allConfigs)
	case "conversational_react":
		return f.createConversationalReactAgent(config, allConfigs)
	default:
		return nil, fmt.Errorf("unsupported Agent type: %s", config.Type)
	}
}

// createZeroShotReactAgent 创建零样本ReAct智能体
func (f *AgentFactory) createZeroShotReactAgent(config *AgentConfig, allConfigs *Config) (*agents.Executor, error) {
	// 获取LLM
	if config.LLMRef == "" {
		return nil, fmt.Errorf("LLM reference is required for zero_shot_react agent")
	}

	llmConfig := allConfigs.LLMs[config.LLMRef]
	llm, err := f.llmFactory.Create(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM for agent: %w", err)
	}

	// 创建零样本智能体
	agent := agents.NewOneShotAgent(
		llm,
		nil,
		agents.WithMaxIterations(f.getMaxSteps(config)),
	)

	// 创建执行器
	executor := agents.NewExecutor(agent)

	return executor, nil
}

// createConversationalReactAgent 创建对话ReAct智能体
func (f *AgentFactory) createConversationalReactAgent(config *AgentConfig, allConfigs *Config) (*agents.Executor, error) {
	// 获取LLM
	if config.LLMRef == "" {
		return nil, fmt.Errorf("LLM reference is required for conversational_react agent")
	}

	llmConfig := allConfigs.LLMs[config.LLMRef]
	llm, err := f.llmFactory.Create(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM for agent: %w", err)
	}

	// 获取记忆
	var memory schema.Memory
	if config.MemoryRef != "" {
		memoryConfig := allConfigs.Memories[config.MemoryRef]
		memory, err = f.memoryFactory.CreateWithLLM(memoryConfig, llm)
		if err != nil {
			return nil, fmt.Errorf("failed to create Memory for agent: %w", err)
		}
	} else {
		// 使用默认记忆 (simple)
		var memErr error
		memory, memErr = f.memoryFactory.createSimple(&MemoryConfig{Type: "simple"})
		if memErr != nil {
			return nil, fmt.Errorf("failed to create default simple memory: %w", memErr)
		}
	}

	// 创建对话智能体
	agent := agents.NewConversationalAgent(
		llm,
		nil,
		agents.WithMemory(memory),
		agents.WithMaxIterations(f.getMaxSteps(config)),
	)

	// 创建执行器
	executor := agents.NewExecutor(agent)

	return executor, nil
}

// createTools 创建工具列表

// getMaxSteps 获取最大步数
func (f *AgentFactory) getMaxSteps(config *AgentConfig) int {
	if config.MaxSteps != nil {
		return *config.MaxSteps
	}
	return 5 // 默认值
}

// CreateWithComponents 使用现有组件创建Agent
func (f *AgentFactory) CreateWithComponents(agentType string, llm llms.Model, memory schema.Memory, tools []tools.Tool, maxSteps int) (*agents.Executor, error) {
	switch agentType {
	case "zero_shot_react":
		agent := agents.NewOneShotAgent(
			llm,
			tools,
			agents.WithMaxIterations(maxSteps),
		)
		return agents.NewExecutor(agent), nil
	case "conversational_react":
		if memory == nil {
			return nil, fmt.Errorf("memory is required for conversational_react agent")
		}
		agent := agents.NewConversationalAgent(
			llm,
			tools,
			agents.WithMemory(memory),
			agents.WithMaxIterations(maxSteps),
		)
		return agents.NewExecutor(agent), nil
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

// CreateExecutorWithAgent 使用现有Agent创建执行器
func (f *AgentFactory) CreateExecutorWithAgent(agent agents.Agent) *agents.Executor {
	return agents.NewExecutor(agent)
}
