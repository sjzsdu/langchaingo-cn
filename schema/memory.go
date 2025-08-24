package schema

import (
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

// MemoryFactory Memory组件工厂
type MemoryFactory struct {
	llmFactory *LLMFactory
}

// NewMemoryFactory 创建Memory工厂实例
func NewMemoryFactory(llmFactory *LLMFactory) *MemoryFactory {
	return &MemoryFactory{
		llmFactory: llmFactory,
	}
}

// Create 根据配置创建Memory实例
func (f *MemoryFactory) Create(config *MemoryConfig, llmConfigs map[string]*LLMConfig) (schema.Memory, error) {
	if config == nil {
		return nil, fmt.Errorf("Memory config is nil")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Memory config: %w", err)
	}

	switch config.Type {
	case "conversation_buffer":
		return f.createConversationBuffer(config)
	case "conversation_token_buffer":
		return f.createConversationTokenBuffer(config, llmConfigs)
	case "simple":
		return f.createSimple(config)
	default:
		return nil, fmt.Errorf("unsupported Memory type: %s", config.Type)
	}
}

// createConversationBuffer 创建ConversationBuffer记忆
func (f *MemoryFactory) createConversationBuffer(config *MemoryConfig) (schema.Memory, error) {
	mem := memory.NewConversationBuffer()

	// 设置消息数量限制
	if config.MaxMessages != nil {
		// ConversationBuffer没有直接的消息数量限制功能
		// 可以考虑使用ConversationWindowBuffer
		return memory.NewConversationWindowBuffer(*config.MaxMessages), nil
	}

	// 设置是否返回消息
	if config.ReturnMessages != nil {
		mem.ReturnMessages = *config.ReturnMessages
	}

	return mem, nil
}

// createConversationSummary 创建ConversationSummary记忆

// createConversationTokenBuffer 创建ConversationTokenBuffer记忆
func (f *MemoryFactory) createConversationTokenBuffer(config *MemoryConfig, llmConfigs map[string]*LLMConfig) (schema.Memory, error) {
	// 需要LLM来计算token数量
	if config.LLMRef == "" {
		return nil, fmt.Errorf("LLM reference is required for conversation_token_buffer memory")
	}

	llmConfig, exists := llmConfigs[config.LLMRef]
	if !exists {
		return nil, fmt.Errorf("referenced LLM '%s' not found", config.LLMRef)
	}

	// 创建LLM实例
	llm, err := f.llmFactory.Create(llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM for memory: %w", err)
	}

	// 设置默认token限制
	maxTokenLimit := 2000
	if config.MaxTokenLimit != nil {
		maxTokenLimit = *config.MaxTokenLimit
	}

	mem := memory.NewConversationTokenBuffer(llm, maxTokenLimit)

	// 设置是否返回消息
	if config.ReturnMessages != nil {
		mem.ReturnMessages = *config.ReturnMessages
	}

	return mem, nil
}

// createSimple 创建Simple记忆
func (f *MemoryFactory) createSimple(config *MemoryConfig) (schema.Memory, error) {
	return memory.NewSimple(), nil
}

// CreateWithLLM 根据配置创建Memory实例（直接传递LLM实例）
func (f *MemoryFactory) CreateWithLLM(config *MemoryConfig, llm llms.Model) (schema.Memory, error) {
	if config == nil {
		return nil, fmt.Errorf("Memory config is nil")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Memory config: %w", err)
	}

	switch config.Type {
	case "conversation_buffer":
		return f.createConversationBuffer(config)
	case "conversation_token_buffer":
		if llm == nil {
			return nil, fmt.Errorf("LLM is required for conversation_token_buffer memory")
		}
		return f.createConversationTokenBufferWithLLM(config, llm)
	case "simple":
		return f.createSimple(config)
	default:
		return nil, fmt.Errorf("unsupported Memory type: %s", config.Type)
	}
}

// createConversationSummaryWithLLM 使用现有LLM创建ConversationSummary记忆

// createConversationTokenBufferWithLLM 使用现有LLM创建ConversationTokenBuffer记忆
func (f *MemoryFactory) createConversationTokenBufferWithLLM(config *MemoryConfig, llm llms.Model) (schema.Memory, error) {
	// 设置默认token限制
	maxTokenLimit := 2000
	if config.MaxTokenLimit != nil {
		maxTokenLimit = *config.MaxTokenLimit
	}

	mem := memory.NewConversationTokenBuffer(llm, maxTokenLimit)

	// 设置是否返回消息
	if config.ReturnMessages != nil {
		mem.ReturnMessages = *config.ReturnMessages
	}

	return mem, nil
}
