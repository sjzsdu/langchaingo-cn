package schema

import (
	"fmt"

	"github.com/tmc/langchaingo/prompts"
)

// PromptFactory Prompt组件工厂
type PromptFactory struct{}

// NewPromptFactory 创建Prompt工厂实例
func NewPromptFactory() *PromptFactory {
	return &PromptFactory{}
}

// Create 根据配置创建Prompt实例
func (f *PromptFactory) Create(config *PromptConfig) (prompts.FormatPrompter, error) {
	if config == nil {
		return nil, fmt.Errorf("Prompt config is nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Prompt config: %w", err)
	}

	switch config.Type {
	case "prompt_template":
		return f.createPromptTemplate(config)
	case "chat_prompt_template":
		return f.createChatPromptTemplate(config)
	default:
		return nil, fmt.Errorf("unsupported Prompt type: %s", config.Type)
	}
}

// createPromptTemplate 创建普通提示模板
func (f *PromptFactory) createPromptTemplate(config *PromptConfig) (prompts.FormatPrompter, error) {
	// 设置输入变量
	inputVariables := config.InputVariables
	if inputVariables == nil {
		inputVariables = []string{}
	}

	// 创建基本模板
	template := prompts.NewPromptTemplate(config.Template, inputVariables)

	// 设置模板格式（如果支持）
	if config.TemplateFormat != "" {
		// PromptTemplate默认支持Go模板格式
		// 可以在这里添加对其他格式的支持
	}

	// 验证模板（如果需要）
	if config.ValidateTemplate != nil && *config.ValidateTemplate {
		// 可以在这里添加模板验证逻辑
	}

	return template, nil
}

// createChatPromptTemplate 创建聊天提示模板
func (f *PromptFactory) createChatPromptTemplate(config *PromptConfig) (prompts.FormatPrompter, error) {
	if len(config.Messages) == 0 {
		return nil, fmt.Errorf("messages are required for chat_prompt_template")
	}

	var messageTemplates []prompts.MessageFormatter
	// 收集输入变量 (简单：用户提供的 + 模板中未解析变量忽略)
	inputVars := config.InputVariables
	if inputVars == nil {
		inputVars = []string{}
	}

	for _, m := range config.Messages {
		switch m.Role {
		case "system":
			messageTemplates = append(messageTemplates, prompts.NewSystemMessagePromptTemplate(m.Template, inputVars))
		case "human", "user":
			messageTemplates = append(messageTemplates, prompts.NewHumanMessagePromptTemplate(m.Template, inputVars))
		case "ai", "assistant":
			messageTemplates = append(messageTemplates, prompts.NewAIMessagePromptTemplate(m.Template, inputVars))
		default:
			return nil, fmt.Errorf("unsupported message role: %s", m.Role)
		}
	}

	chatTemplate := prompts.NewChatPromptTemplate(messageTemplates)
	return chatTemplate, nil
}

// extractVariablesFromTemplate 从模板中提取变量（简单实现）
// 这个函数假设使用Go模板语法 {{.variable}}
func extractVariablesFromTemplate(template string) []string {
	// 简单的正则匹配提取变量
	// 在实际应用中，可以使用更复杂的解析逻辑
	var variables []string

	// 这里可以实现更复杂的变量提取逻辑
	// 暂时返回空切片，让用户手动指定输入变量

	return variables
}

// CreateWithSystemMessage 创建带有系统消息的聊天模板
func (f *PromptFactory) CreateWithSystemMessage(systemMsg, humanMsg string, inputVars []string) (prompts.FormatPrompter, error) {
	messages := []prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate(systemMsg, inputVars),
		prompts.NewHumanMessagePromptTemplate(humanMsg, inputVars),
	}
	tpl := prompts.NewChatPromptTemplate(messages)
	return tpl, nil
}

// CreateSimpleTemplate 创建简单模板
func (f *PromptFactory) CreateSimpleTemplate(template string, inputVars []string) prompts.FormatPrompter {
	return prompts.NewPromptTemplate(template, inputVars)
}
