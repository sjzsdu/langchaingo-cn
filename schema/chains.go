package schema

import (
	"fmt"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// ChainFactory Chain组件工厂
type ChainFactory struct {
	llmFactory    *LLMFactory
	memoryFactory *MemoryFactory
	promptFactory *PromptFactory
}

// NewChainFactory 创建Chain工厂实例
func NewChainFactory(llmFactory *LLMFactory, memoryFactory *MemoryFactory, promptFactory *PromptFactory) *ChainFactory {
	return &ChainFactory{
		llmFactory:    llmFactory,
		memoryFactory: memoryFactory,
		promptFactory: promptFactory,
	}
}

// Create 根据配置创建Chain实例
func (f *ChainFactory) Create(config *ChainConfig, allConfigs *Config) (chains.Chain, error) {
	if config == nil {
		return nil, fmt.Errorf("Chain config is nil")
	}

	// 验证配置
	if err := config.ValidateReferences(allConfigs); err != nil {
		return nil, fmt.Errorf("invalid Chain config: %w", err)
	}

	switch config.Type {
	case "llm":
		return f.createLLMChain(config, allConfigs)
	case "conversation":
		return f.createConversationChain(config, allConfigs)
	case "sequential":
		return f.createSequentialChain(config, allConfigs)
	case "stuff_documents":
		return f.createStuffDocumentsChain(config, allConfigs)
	case "map_reduce":
		return f.createMapReduceChain(config, allConfigs)
	default:
		return nil, fmt.Errorf("unsupported Chain type: %s", config.Type)
	}
}

// createLLMChain 创建LLM链
func (f *ChainFactory) createLLMChain(config *ChainConfig, allConfigs *Config) (chains.Chain, error) {
	// 获取LLM
	if config.LLMRef == "" {
		return nil, fmt.Errorf("LLM reference is required for llm chain")
	}

	llmConfig := allConfigs.LLMs[config.LLMRef]
	llm, err := f.llmFactory.Create(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM for chain: %w", err)
	}

	// 获取Prompt
	var prompt prompts.FormatPrompter
	if config.PromptRef != "" {
		promptConfig := allConfigs.Prompts[config.PromptRef]
		prompt, err = f.promptFactory.Create(promptConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Prompt for chain: %w", err)
		}
	} else {
		// 使用默认的简单模板
		prompt = f.promptFactory.CreateSimpleTemplate("{{.input}}", []string{"input"})
	}

	// 创建LLM链
	chain := chains.NewLLMChain(llm, prompt)

	return chain, nil
}

// createConversationChain 创建对话链
func (f *ChainFactory) createConversationChain(config *ChainConfig, allConfigs *Config) (chains.Chain, error) {
	// 获取LLM
	if config.LLMRef == "" {
		return nil, fmt.Errorf("LLM reference is required for conversation chain")
	}

	llmConfig := allConfigs.LLMs[config.LLMRef]
	llm, err := f.llmFactory.Create(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM for chain: %w", err)
	}

	// 获取Memory
	var memory schema.Memory
	if config.MemoryRef != "" {
		memoryConfig := allConfigs.Memories[config.MemoryRef]
		memory, err = f.memoryFactory.CreateWithLLM(memoryConfig, llm)
		if err != nil {
			return nil, fmt.Errorf("failed to create Memory for chain: %w", err)
		}
	} else {
		// 使用默认的简单记忆
		var memErr error
		memory, memErr = f.memoryFactory.createSimple(&MemoryConfig{Type: "simple"})
		if memErr != nil {
			return nil, fmt.Errorf("failed to create default simple memory: %w", memErr)
		}
	}

	// 创建对话链
	chain := chains.NewConversation(llm, memory)

	return chain, nil
}

// createSequentialChain 创建顺序链
func (f *ChainFactory) createSequentialChain(config *ChainConfig, allConfigs *Config) (chains.Chain, error) {
	if len(config.Chains) == 0 {
		return nil, fmt.Errorf("sub-chains are required for sequential chain")
	}

	// 创建子链
	var subChains []chains.Chain
	for _, chainRef := range config.Chains {
		chainConfig, exists := allConfigs.Chains[chainRef]
		if !exists {
			return nil, fmt.Errorf("referenced chain '%s' not found", chainRef)
		}

		subChain, err := f.Create(chainConfig, allConfigs)
		if err != nil {
			return nil, fmt.Errorf("failed to create sub-chain '%s': %w", chainRef, err)
		}

		subChains = append(subChains, subChain)
	}

	// 创建顺序链
	if len(config.InputKeys) > 0 && len(config.OutputKeys) > 0 {
		// 使用完整的顺序链
		chain, err := chains.NewSequentialChain(subChains, config.InputKeys, config.OutputKeys)
		if err != nil {
			return nil, fmt.Errorf("failed to create sequential chain: %w", err)
		}
		return chain, nil
	} else {
		// 使用简单顺序链
		chain, err := chains.NewSimpleSequentialChain(subChains)
		if err != nil {
			return nil, fmt.Errorf("failed to create simple sequential chain: %w", err)
		}
		return chain, nil
	}
}

// createStuffDocumentsChain 创建文档填充链
func (f *ChainFactory) createStuffDocumentsChain(config *ChainConfig, allConfigs *Config) (chains.Chain, error) {
	// 先创建LLM链
	llmChain, err := f.createLLMChain(config, allConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM chain for stuff documents: %w", err)
	}

	// 转换为LLMChain类型
	llmChainTyped, ok := llmChain.(*chains.LLMChain)
	if !ok {
		return nil, fmt.Errorf("expected LLMChain, got %T", llmChain)
	}

	// 创建文档填充链
	stuffChain := chains.NewStuffDocuments(llmChainTyped)

	// 设置分隔符
	if config.Separator != "" {
		stuffChain.Separator = config.Separator
	}

	return stuffChain, nil
}

// createMapReduceChain 创建MapReduce链
func (f *ChainFactory) createMapReduceChain(config *ChainConfig, allConfigs *Config) (chains.Chain, error) {
	// 需要两个子链：map链和reduce链
	if len(config.Chains) < 2 {
		return nil, fmt.Errorf("map-reduce chain requires at least 2 sub-chains (map and reduce)")
	}

	// 创建map链
	mapChainRef := config.Chains[0]
	mapChainConfig, exists := allConfigs.Chains[mapChainRef]
	if !exists {
		return nil, fmt.Errorf("referenced map chain '%s' not found", mapChainRef)
	}

	mapChain, err := f.createLLMChain(mapChainConfig, allConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to create map chain: %w", err)
	}

	mapChainTyped, ok := mapChain.(*chains.LLMChain)
	if !ok {
		return nil, fmt.Errorf("map chain must be LLMChain, got %T", mapChain)
	}

	// 创建reduce链
	reduceChainRef := config.Chains[1]
	reduceChainConfig, exists := allConfigs.Chains[reduceChainRef]
	if !exists {
		return nil, fmt.Errorf("referenced reduce chain '%s' not found", reduceChainRef)
	}

	reduceChain, err := f.Create(reduceChainConfig, allConfigs)
	if err != nil {
		return nil, fmt.Errorf("failed to create reduce chain: %w", err)
	}

	// 创建MapReduce链
	mapReduceChain := chains.NewMapReduceDocuments(mapChainTyped, reduceChain)

	// MapReduceDocuments 当前未暴露并发配置字段，忽略 MaxConcurrency 设置

	return mapReduceChain, nil
}

// CreateWithComponents 使用现有组件创建Chain
func (f *ChainFactory) CreateWithComponents(chainType string, llm llms.Model, memory schema.Memory, prompt prompts.FormatPrompter) (chains.Chain, error) {
	switch chainType {
	case "llm":
		return chains.NewLLMChain(llm, prompt), nil
	case "conversation":
		if memory == nil {
			return nil, fmt.Errorf("memory is required for conversation chain")
		}
		return chains.NewConversation(llm, memory), nil
	default:
		return nil, fmt.Errorf("unsupported chain type for direct creation: %s", chainType)
	}
}
