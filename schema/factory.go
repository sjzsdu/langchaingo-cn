package schema

import (
	"fmt"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// Factory 统一的组件工厂管理器
type Factory struct {
	llmFactory       *LLMFactory
	memoryFactory    *MemoryFactory
	promptFactory    *PromptFactory
	embeddingFactory *EmbeddingFactory
	chainFactory     *ChainFactory
	agentFactory     *AgentFactory
}

// NewFactory 创建新的统一工厂管理器
func NewFactory() *Factory {
	llmFactory := NewLLMFactory()
	memoryFactory := NewMemoryFactory(llmFactory)
	promptFactory := NewPromptFactory()
	embeddingFactory := NewEmbeddingFactory()
	chainFactory := NewChainFactory(llmFactory, memoryFactory, promptFactory)
	agentFactory := NewAgentFactory(llmFactory, memoryFactory)

	return &Factory{
		llmFactory:       llmFactory,
		memoryFactory:    memoryFactory,
		promptFactory:    promptFactory,
		embeddingFactory: embeddingFactory,
		chainFactory:     chainFactory,
		agentFactory:     agentFactory,
	}
}

// Application 应用程序实例，包含所有创建的组件
type Application struct {
	LLMs       map[string]llms.Model
	Memories   map[string]schema.Memory
	Prompts    map[string]prompts.FormatPrompter
	Embeddings map[string]embeddings.Embedder
	Chains     map[string]chains.Chain
	Agents     map[string]*agents.Executor
}

// CreateApplication 根据配置创建完整的应用程序
func (f *Factory) CreateApplication(config *Config) (*Application, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	app := &Application{
		LLMs:       make(map[string]llms.Model),
		Memories:   make(map[string]schema.Memory),
		Prompts:    make(map[string]prompts.FormatPrompter),
		Embeddings: make(map[string]embeddings.Embedder),
		Chains:     make(map[string]chains.Chain),
		Agents:     make(map[string]*agents.Executor),
	}

	// 创建LLM组件
	for name, llmConfig := range config.LLMs {
		llm, err := f.llmFactory.Create(llmConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create LLM '%s': %w", name, err)
		}
		app.LLMs[name] = llm
	}

	// 创建Memory组件
	for name, memoryConfig := range config.Memories {
		memory, err := f.memoryFactory.Create(memoryConfig, config.LLMs)
		if err != nil {
			return nil, fmt.Errorf("failed to create Memory '%s': %w", name, err)
		}
		app.Memories[name] = memory
	}

	// 创建Prompt组件
	for name, promptConfig := range config.Prompts {
		prompt, err := f.promptFactory.Create(promptConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Prompt '%s': %w", name, err)
		}
		app.Prompts[name] = prompt
	}

	// 创建Embedding组件
	for name, embeddingConfig := range config.Embeddings {
		embedding, err := f.embeddingFactory.Create(embeddingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Embedding '%s': %w", name, err)
		}
		app.Embeddings[name] = embedding
	}

	// 创建Chain组件
	for name, chainConfig := range config.Chains {
		chain, err := f.chainFactory.Create(chainConfig, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Chain '%s': %w", name, err)
		}
		app.Chains[name] = chain
	}

	// 创建Agent组件
	for name, agentConfig := range config.Agents {
		agent, err := f.agentFactory.Create(agentConfig, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Agent '%s': %w", name, err)
		}
		app.Agents[name] = agent
	}

	return app, nil
}

// CreateLLM 创建单个LLM组件
func (f *Factory) CreateLLM(config *LLMConfig) (llms.Model, error) {
	return f.llmFactory.Create(config)
}

// CreateMemory 创建单个Memory组件
func (f *Factory) CreateMemory(config *MemoryConfig, llmConfigs map[string]*LLMConfig) (schema.Memory, error) {
	return f.memoryFactory.Create(config, llmConfigs)
}

// CreateMemoryWithLLM 使用现有LLM创建单个Memory组件
func (f *Factory) CreateMemoryWithLLM(config *MemoryConfig, llm llms.Model) (schema.Memory, error) {
	return f.memoryFactory.CreateWithLLM(config, llm)
}

// CreatePrompt 创建单个Prompt组件
func (f *Factory) CreatePrompt(config *PromptConfig) (prompts.FormatPrompter, error) {
	return f.promptFactory.Create(config)
}

// CreateEmbedding 创建单个Embedding组件
func (f *Factory) CreateEmbedding(config *EmbeddingConfig) (embeddings.Embedder, error) {
	return f.embeddingFactory.Create(config)
}

// CreateChain 创建单个Chain组件
func (f *Factory) CreateChain(config *ChainConfig, allConfigs *Config) (chains.Chain, error) {
	return f.chainFactory.Create(config, allConfigs)
}

// CreateAgent 创建单个Agent组件
func (f *Factory) CreateAgent(config *AgentConfig, allConfigs *Config) (*agents.Executor, error) {
	return f.agentFactory.Create(config, allConfigs)
}

// GetLLMFactory 获取LLM工厂
func (f *Factory) GetLLMFactory() *LLMFactory {
	return f.llmFactory
}

// GetMemoryFactory 获取Memory工厂
func (f *Factory) GetMemoryFactory() *MemoryFactory {
	return f.memoryFactory
}

// GetPromptFactory 获取Prompt工厂
func (f *Factory) GetPromptFactory() *PromptFactory {
	return f.promptFactory
}

// GetEmbeddingFactory 获取Embedding工厂
func (f *Factory) GetEmbeddingFactory() *EmbeddingFactory {
	return f.embeddingFactory
}

// GetChainFactory 获取Chain工厂
func (f *Factory) GetChainFactory() *ChainFactory {
	return f.chainFactory
}

// GetAgentFactory 获取Agent工厂
func (f *Factory) GetAgentFactory() *AgentFactory {
	return f.agentFactory
}

// 全局工厂实例
var globalFactory *Factory

// init 初始化全局工厂
func init() {
	globalFactory = NewFactory()
}

// GetGlobalFactory 获取全局工厂实例
func GetGlobalFactory() *Factory {
	return globalFactory
}

// CreateApplicationFromFile 从文件创建应用程序
func CreateApplicationFromFile(filename string) (*Application, error) {
	config, err := LoadConfigFromFile(filename)
	if err != nil {
		return nil, err
	}
	return globalFactory.CreateApplication(config)
}

// CreateApplicationFromJSON 从JSON字符串创建应用程庋
func CreateApplicationFromJSON(jsonStr string) (*Application, error) {
	config, err := LoadConfigFromJSON(jsonStr)
	if err != nil {
		return nil, err
	}
	return globalFactory.CreateApplication(config)
}

// 便利函数，使用全局工厂创建单个组件

// CreateLLMFromConfig 使用全局工厂创建LLM
func CreateLLMFromConfig(config *LLMConfig) (llms.Model, error) {
	return globalFactory.CreateLLM(config)
}

// CreateMemoryFromConfig 使用全局工厂创建Memory
func CreateMemoryFromConfig(config *MemoryConfig, llmConfigs map[string]*LLMConfig) (schema.Memory, error) {
	return globalFactory.CreateMemory(config, llmConfigs)
}

// CreatePromptFromConfig 使用全局工厂创建Prompt
func CreatePromptFromConfig(config *PromptConfig) (prompts.FormatPrompter, error) {
	return globalFactory.CreatePrompt(config)
}

// CreateEmbeddingFromConfig 使用全局工厂创建Embedding
func CreateEmbeddingFromConfig(config *EmbeddingConfig) (embeddings.Embedder, error) {
	return globalFactory.CreateEmbedding(config)
}

// CreateChainFromConfig 使用全局工厂创建Chain
func CreateChainFromConfig(config *ChainConfig, allConfigs *Config) (chains.Chain, error) {
	return globalFactory.CreateChain(config, allConfigs)
}

// CreateAgentFromConfig 使用全局工厂创建Agent
func CreateAgentFromConfig(config *AgentConfig, allConfigs *Config) (*agents.Executor, error) {
	return globalFactory.CreateAgent(config, allConfigs)
}
