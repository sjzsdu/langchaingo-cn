package schema

import (
	"fmt"
	"strings"
)

// SchemaError schema包的错误类型
type SchemaError struct {
	Type    string // 错误类型：validation, creation, configuration
	Path    string // 错误路径，如 "llms.openai", "chains.conversation"
	Message string // 错误消息
	Cause   error  // 原始错误
}

// Error 实现error接口
func (e *SchemaError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Type, e.Path, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap 返回原始错误
func (e *SchemaError) Unwrap() error {
	return e.Cause
}

// 错误类型常量
const (
	ErrorTypeValidation    = "validation"
	ErrorTypeCreation      = "creation"
	ErrorTypeConfiguration = "configuration"
	ErrorTypeReference     = "reference"
	ErrorTypeUnsupported   = "unsupported"
)

// NewValidationError 创建验证错误
func NewValidationError(path, message string, cause error) *SchemaError {
	return &SchemaError{
		Type:    ErrorTypeValidation,
		Path:    path,
		Message: message,
		Cause:   cause,
	}
}

// NewCreationError 创建创建错误
func NewCreationError(path, message string, cause error) *SchemaError {
	return &SchemaError{
		Type:    ErrorTypeCreation,
		Path:    path,
		Message: message,
		Cause:   cause,
	}
}

// NewConfigurationError 创建配置错误
func NewConfigurationError(path, message string, cause error) *SchemaError {
	return &SchemaError{
		Type:    ErrorTypeConfiguration,
		Path:    path,
		Message: message,
		Cause:   cause,
	}
}

// NewReferenceError 创建引用错误
func NewReferenceError(path, message string, cause error) *SchemaError {
	return &SchemaError{
		Type:    ErrorTypeReference,
		Path:    path,
		Message: message,
		Cause:   cause,
	}
}

// NewUnsupportedError 创建不支持错误
func NewUnsupportedError(path, message string, cause error) *SchemaError {
	return &SchemaError{
		Type:    ErrorTypeUnsupported,
		Path:    path,
		Message: message,
		Cause:   cause,
	}
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool            // 是否有效
	Errors []*SchemaError  // 错误列表
	Warnings []string      // 警告列表
}

// AddError 添加错误
func (vr *ValidationResult) AddError(err *SchemaError) {
	vr.Errors = append(vr.Errors, err)
	vr.Valid = false
}

// AddWarning 添加警告
func (vr *ValidationResult) AddWarning(warning string) {
	vr.Warnings = append(vr.Warnings, warning)
}

// HasErrors 检查是否有错误
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings 检查是否有警告
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// GetErrorMessages 获取所有错误消息
func (vr *ValidationResult) GetErrorMessages() []string {
	messages := make([]string, len(vr.Errors))
	for i, err := range vr.Errors {
		messages[i] = err.Error()
	}
	return messages
}

// String 返回验证结果的字符串表示
func (vr *ValidationResult) String() string {
	var sb strings.Builder

	if vr.Valid {
		sb.WriteString("Validation: PASSED")
	} else {
		sb.WriteString("Validation: FAILED")
	}

	if len(vr.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("\nErrors (%d):", len(vr.Errors)))
		for _, err := range vr.Errors {
			sb.WriteString(fmt.Sprintf("\n  - %s", err.Error()))
		}
	}

	if len(vr.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("\nWarnings (%d):", len(vr.Warnings)))
		for _, warning := range vr.Warnings {
			sb.WriteString(fmt.Sprintf("\n  - %s", warning))
		}
	}

	return sb.String()
}

// ValidateConfig 全面验证配置
func ValidateConfig(config *Config) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if config == nil {
		result.AddError(NewValidationError("", "config is nil", nil))
		return result
	}

	// 验证LLM配置
	for name, llmConfig := range config.LLMs {
		if err := llmConfig.Validate(); err != nil {
			result.AddError(NewValidationError(fmt.Sprintf("llms.%s", name), err.Error(), err))
		}

		// 检查API密钥
		if llmConfig.APIKey == "" {
			if getDefaultAPIKeyForType(llmConfig.Type) == "" {
				result.AddWarning(fmt.Sprintf("LLM '%s': API key not configured and no default environment variable found", name))
			}
		}
	}

	// 验证Memory配置
	for name, memoryConfig := range config.Memories {
		if err := memoryConfig.Validate(); err != nil {
			result.AddError(NewValidationError(fmt.Sprintf("memories.%s", name), err.Error(), err))
		}

		// 检查LLM引用
		if memoryConfig.LLMRef != "" {
			if _, exists := config.LLMs[memoryConfig.LLMRef]; !exists {
				result.AddError(NewReferenceError(fmt.Sprintf("memories.%s", name), 
					fmt.Sprintf("referenced LLM '%s' not found", memoryConfig.LLMRef), nil))
			}
		}
	}

	// 验证Prompt配置
	for name, promptConfig := range config.Prompts {
		if err := promptConfig.Validate(); err != nil {
			result.AddError(NewValidationError(fmt.Sprintf("prompts.%s", name), err.Error(), err))
		}
	}

	// 验证Embedding配置
	for name, embeddingConfig := range config.Embeddings {
		if err := embeddingConfig.Validate(); err != nil {
			result.AddError(NewValidationError(fmt.Sprintf("embeddings.%s", name), err.Error(), err))
		}

		// 检查API密钥
		if embeddingConfig.APIKey == "" {
			if getDefaultAPIKeyForEmbedding(embeddingConfig.Type) == "" {
				result.AddWarning(fmt.Sprintf("Embedding '%s': API key not configured and no default environment variable found", name))
			}
		}
	}

	// 验证Chain配置
	for name, chainConfig := range config.Chains {
		if err := chainConfig.ValidateReferences(config); err != nil {
			result.AddError(NewValidationError(fmt.Sprintf("chains.%s", name), err.Error(), err))
		}
	}

	// 验证Agent配置
	for name, agentConfig := range config.Agents {
		if err := agentConfig.ValidateReferences(config); err != nil {
			result.AddError(NewValidationError(fmt.Sprintf("agents.%s", name), err.Error(), err))
		}
	}

	// 检查循环引用
	if cyclicRefs := detectCyclicReferences(config); len(cyclicRefs) > 0 {
		for _, ref := range cyclicRefs {
			result.AddError(NewReferenceError(ref, "cyclic reference detected", nil))
		}
	}

	return result
}

// getDefaultAPIKeyForType 获取默认API密钥环境变量
func getDefaultAPIKeyForType(llmType string) string {
	switch llmType {
	case "openai":
		return "OPENAI_API_KEY"
	case "deepseek":
		return "DEEPSEEK_API_KEY"
	case "kimi":
		return "KIMI_API_KEY"
	case "qwen":
		return "QWEN_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	default:
		return ""
	}
}

// getDefaultAPIKeyForEmbedding 获取Embedding默认API密钥环境变量
func getDefaultAPIKeyForEmbedding(embeddingType string) string {
	switch embeddingType {
	case "openai":
		return "OPENAI_API_KEY"
	case "voyage":
		return "VOYAGEAI_API_KEY"
	case "cohere":
		return "COHERE_API_KEY"
	default:
		return ""
	}
}

// detectCyclicReferences 检测循环引用
func detectCyclicReferences(config *Config) []string {
	var cyclicRefs []string

	// 检查Chain中的循环引用
	for name, chainConfig := range config.Chains {
		if chainConfig.Type == "sequential" {
			visited := make(map[string]bool)
			if hasCycle := checkChainCycle(name, chainConfig, config.Chains, visited, make(map[string]bool)); hasCycle {
				cyclicRefs = append(cyclicRefs, fmt.Sprintf("chains.%s", name))
			}
		}
	}

	return cyclicRefs
}

// checkChainCycle 检查链中的循环
func checkChainCycle(chainName string, chainConfig *ChainConfig, allChains map[string]*ChainConfig, visited, recursionStack map[string]bool) bool {
	if recursionStack[chainName] {
		return true // 找到循环
	}

	if visited[chainName] {
		return false // 已经检查过，没有循环
	}

	visited[chainName] = true
	recursionStack[chainName] = true

	// 检查子链
	for _, subChainName := range chainConfig.Chains {
		if subChainConfig, exists := allChains[subChainName]; exists {
			if checkChainCycle(subChainName, subChainConfig, allChains, visited, recursionStack) {
				return true
			}
		}
	}

	recursionStack[chainName] = false
	return false
}