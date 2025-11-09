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
	chainFactory *ChainFactory
}

// NewAgentFactory 创建Agent工厂实例
func NewAgentFactory(chainFactory *ChainFactory) *AgentFactory {
	return &AgentFactory{
		chainFactory: chainFactory,
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
	// 获取Chain配置
	if config.ChainRef == "" {
		return nil, fmt.Errorf("chain reference is required for zero_shot_react agent")
	}

	chainConfig := allConfigs.Chains[config.ChainRef]

	// 直接从Chain配置中获取LLM，而不是从Chain对象中提取
	if chainConfig.LLMRef == "" {
		return nil, fmt.Errorf("LLM reference is required in chain for zero_shot_react agent")
	}

	llmConfig := allConfigs.LLMs[chainConfig.LLMRef]
	llm, err := f.chainFactory.llmFactory.Create(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM for agent: %w", err)
	}

	// 从Chain配置中获取工具(如果有)
	tools := f.getToolsFromOptions(config.Options)

	// 创建零样本智能体
	agent := agents.NewOneShotAgent(
		llm, // 直接使用LLM
		tools,
		agents.WithMaxIterations(f.getMaxIterations(config)),
		agents.WithOutputKey(f.getOutputKey(config)),
	)

	// 创建执行器
	executor := agents.NewExecutor(agent)

	return executor, nil
}

// createConversationalReactAgent 创建对话ReAct智能体
func (f *AgentFactory) createConversationalReactAgent(config *AgentConfig, allConfigs *Config) (*agents.Executor, error) {
	// 获取Chain配置
	if config.ChainRef == "" {
		return nil, fmt.Errorf("chain reference is required for conversational_react agent")
	}

	chainConfig := allConfigs.Chains[config.ChainRef]

	// 直接从Chain配置中获取LLM，而不是从Chain对象中提取
	if chainConfig.LLMRef == "" {
		return nil, fmt.Errorf("LLM reference is required in chain for conversational_react agent")
	}

	llmConfig := allConfigs.LLMs[chainConfig.LLMRef]
	llm, err := f.chainFactory.llmFactory.Create(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM for agent: %w", err)
	}

	// 从Chain配置中获取工具(如果有)
	tools := f.getToolsFromOptions(config.Options)

	// 创建对话智能体
	agent := agents.NewConversationalAgent(
		llm, // 直接使用LLM
		tools,
		agents.WithMaxIterations(f.getMaxIterations(config)),
		agents.WithOutputKey(f.getOutputKey(config)),
	)

	// 创建执行器
	executor := agents.NewExecutor(agent)

	return executor, nil
}

// getToolsFromOptions 从Options中获取工具列表
func (f *AgentFactory) getToolsFromOptions(options map[string]interface{}) []tools.Tool {
	if options == nil {
		return nil
	}

	// 这里可以根据options中的配置创建工具
	// 目前返回空切片，实际实现中可以根据需要添加工具解析逻辑
	return []tools.Tool{}
}

// getMaxIterations 获取最大迭代次数
func (f *AgentFactory) getMaxIterations(config *AgentConfig) int {
	// 可以从Options中获取，或使用默认值
	if maxIter, ok := config.Options["max_iterations"].(int); ok {
		return maxIter
	}
	return 5 // 默认值
}

// getOutputKey 获取输出键
func (f *AgentFactory) getOutputKey(config *AgentConfig) string {
	if config.OutputKey != "" {
		return config.OutputKey
	}
	return "output" // 默认值
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
