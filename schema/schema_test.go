package schema

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromJSON(t *testing.T) {
	jsonConfig := `{
		"llms": {
			"test_llm": {
				"type": "openai",
				"model": "gpt-3.5-turbo",
				"api_key": "test-key"
			}
		}
	}`

	config, err := LoadConfigFromJSON(jsonConfig)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Len(t, config.LLMs, 1)
	assert.Equal(t, "openai", config.LLMs["test_llm"].Type)
	assert.Equal(t, "gpt-3.5-turbo", config.LLMs["test_llm"].Model)
	assert.Equal(t, "test-key", config.LLMs["test_llm"].APIKey)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectValid bool
		expectError string
	}{
		{
			name: "valid basic config",
			config: &Config{
				LLMs: map[string]*LLMConfig{
					"test": {
						Type:   "openai",
						Model:  "gpt-3.5-turbo",
						APIKey: "test-key",
					},
				},
			},
			expectValid: true,
		},
		{
			name: "invalid llm type",
			config: &Config{
				LLMs: map[string]*LLMConfig{
					"test": {
						Type:   "invalid",
						Model:  "gpt-3.5-turbo",
						APIKey: "test-key",
					},
				},
			},
			expectValid: false,
			expectError: "unsupported type",
		},
		{
			name: "missing model",
			config: &Config{
				LLMs: map[string]*LLMConfig{
					"test": {
						Type:   "openai",
						APIKey: "test-key",
					},
				},
			},
			expectValid: false,
			expectError: "model is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.expectError != "" {
					assert.Contains(t, err.Error(), tt.expectError)
				}
			}
		})
	}
}

func TestLLMFactory(t *testing.T) {
	factory := NewLLMFactory()

	t.Run("create openai llm", func(t *testing.T) {
		config := &LLMConfig{
			Type:        "openai",
			Model:       "gpt-3.5-turbo",
			APIKey:      "test-key",
			Temperature: floatPtr(0.7),
			MaxTokens:   intPtr(1000),
		}

		llm, err := factory.Create(config)
		assert.NoError(t, err)
		assert.NotNil(t, llm)
	})

	t.Run("invalid config", func(t *testing.T) {
		config := &LLMConfig{
			Type: "invalid",
		}

		_, err := factory.Create(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})
}

func TestMemoryFactory(t *testing.T) {
	llmFactory := NewLLMFactory()
	memoryFactory := NewMemoryFactory(llmFactory)

	t.Run("create conversation buffer", func(t *testing.T) {
		config := &MemoryConfig{
			Type:           "conversation_buffer",
			MaxMessages:    intPtr(10),
			ReturnMessages: boolPtr(true),
		}

		memory, err := memoryFactory.Create(config, nil)
		assert.NoError(t, err)
		assert.NotNil(t, memory)
	})

	t.Run("create simple memory", func(t *testing.T) {
		config := &MemoryConfig{
			Type: "simple",
		}

		memory, err := memoryFactory.Create(config, nil)
		assert.NoError(t, err)
		assert.NotNil(t, memory)
	})
}

func TestPromptFactory(t *testing.T) {
	factory := NewPromptFactory()

	t.Run("create prompt template", func(t *testing.T) {
		config := &PromptConfig{
			Type:           "prompt_template",
			Template:       "Hello {{.name}}!",
			InputVariables: []string{"name"},
		}

		prompt, err := factory.Create(config)
		assert.NoError(t, err)
		assert.NotNil(t, prompt)
	})

	t.Run("create chat prompt template", func(t *testing.T) {
		config := &PromptConfig{
			Type: "chat_prompt_template",
			Messages: []ChatMessageConfig{
				{
					Role:     "system",
					Template: "You are a helpful assistant.",
				},
				{
					Role:     "human",
					Template: "{{.question}}",
				},
			},
			InputVariables: []string{"question"},
		}

		prompt, err := factory.Create(config)
		assert.NoError(t, err)
		assert.NotNil(t, prompt)
	})
}

func TestEmbeddingFactory(t *testing.T) {
	factory := NewEmbeddingFactory()

	t.Run("create openai embedding", func(t *testing.T) {
		config := &EmbeddingConfig{
			Type:      "openai",
			Model:     "text-embedding-ada-002",
			APIKey:    "test-key",
			BatchSize: intPtr(100),
		}

		embedder, err := factory.Create(config)
		assert.NoError(t, err)
		assert.NotNil(t, embedder)
	})
}

func TestCompleteApplicationCreation(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	jsonConfig := `{
		"llms": {
			"main_llm": {
				"type": "openai",
				"model": "gpt-3.5-turbo",
				"api_key": "${OPENAI_API_KEY}"
			}
		},
		"memories": {
			"chat_memory": {
				"type": "conversation_buffer",
				"max_messages": 10
			}
		},
		"prompts": {
			"chat_prompt": {
				"type": "prompt_template",
				"template": "Answer: {{.input}}",
				"input_variables": ["input"]
			}
		},
		"chains": {
			"main_chain": {
				"type": "llm",
				"llm_ref": "main_llm",
				"prompt_ref": "chat_prompt"
			}
		}
	}`

	factory := NewFactory()
	config, err := LoadConfigFromJSON(jsonConfig)
	require.NoError(t, err)

	app, err := factory.CreateApplication(config)
	assert.NoError(t, err)
	assert.NotNil(t, app)

	// 验证组件都已创建
	assert.Len(t, app.LLMs, 1)
	assert.Len(t, app.Memories, 1)
	assert.Len(t, app.Prompts, 1)
	assert.Len(t, app.Chains, 1)

	// 验证组件可以正常访问
	assert.NotNil(t, app.LLMs["main_llm"])
	assert.NotNil(t, app.Memories["chat_memory"])
	assert.NotNil(t, app.Prompts["chat_prompt"])
	assert.NotNil(t, app.Chains["main_chain"])
}

func TestEnvironmentVariableExpansion(t *testing.T) {
	os.Setenv("TEST_API_KEY", "secret-key-123")
	defer os.Unsetenv("TEST_API_KEY")

	jsonConfig := `{
		"llms": {
			"test_llm": {
				"type": "openai",
				"model": "gpt-3.5-turbo", 
				"api_key": "${TEST_API_KEY}"
			}
		}
	}`

	config, err := LoadConfigFromJSON(jsonConfig)
	require.NoError(t, err)

	assert.Equal(t, "secret-key-123", config.LLMs["test_llm"].APIKey)
}

func TestValidationResult(t *testing.T) {
	result := &ValidationResult{Valid: true}

	// 添加错误
	result.AddError(NewValidationError("test.path", "test error", nil))
	assert.False(t, result.Valid)
	assert.True(t, result.HasErrors())
	assert.Len(t, result.Errors, 1)

	// 添加警告
	result.AddWarning("test warning")
	assert.True(t, result.HasWarnings())
	assert.Len(t, result.Warnings, 1)

	// 获取错误消息
	messages := result.GetErrorMessages()
	assert.Len(t, messages, 1)
	assert.Contains(t, messages[0], "test error")
}

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
