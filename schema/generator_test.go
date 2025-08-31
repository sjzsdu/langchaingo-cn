package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigGenerator(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	generator := NewConfigGenerator(tempDir)

	t.Run("生成LLM配置", func(t *testing.T) {
		template := LLMTemplate{
			Type:        "deepseek",
			Model:       "deepseek-chat",
			Temperature: 0.7,
			MaxTokens:   2048,
		}

		filename := "test_llm.json"
		err := generator.GenerateLLMConfig(template, filename)
		require.NoError(t, err)

		// 验证文件是否生成
		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)

		// 验证内容
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var config Config
		err = json.Unmarshal(data, &config)
		require.NoError(t, err)

		assert.Len(t, config.LLMs, 1)
		llm := config.LLMs["main_llm"]
		assert.Equal(t, "deepseek", llm.Type)
		assert.Equal(t, "deepseek-chat", llm.Model)
		assert.Equal(t, "${DEEPSEEK_API_KEY}", llm.APIKey)
		assert.Equal(t, 0.7, *llm.Temperature)
		assert.Equal(t, 2048, *llm.MaxTokens)
	})

	t.Run("生成Chain配置", func(t *testing.T) {
		template := ChainTemplate{
			Type: "conversation",
			LLMTemplate: LLMTemplate{
				Type:        "kimi",
				Model:       "moonshot-v1-8k",
				Temperature: 0.5,
			},
			MemoryType:     "conversation_buffer",
			PromptTemplate: "你好，{{.input}}",
			InputVariables: []string{"input"},
		}

		filename := "test_chain.json"
		err := generator.GenerateChainConfig(template, filename)
		require.NoError(t, err)

		// 验证文件是否生成
		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)

		// 验证内容
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var config Config
		err = json.Unmarshal(data, &config)
		require.NoError(t, err)

		// 验证LLM
		assert.Len(t, config.LLMs, 1)
		llm := config.LLMs["chain_llm"]
		assert.Equal(t, "kimi", llm.Type)
		assert.Equal(t, "moonshot-v1-8k", llm.Model)

		// 验证Memory
		assert.Len(t, config.Memories, 1)
		memory := config.Memories["chain_memory"]
		assert.Equal(t, "conversation_buffer", memory.Type)

		// 验证Prompt
		assert.Len(t, config.Prompts, 1)
		prompt := config.Prompts["chain_prompt"]
		assert.Equal(t, "prompt_template", prompt.Type)
		assert.Equal(t, "你好，{{.input}}", prompt.Template)

		// 验证Chain
		assert.Len(t, config.Chains, 1)
		chain := config.Chains["main_chain"]
		assert.Equal(t, "conversation", chain.Type)
		assert.Equal(t, "chain_llm", chain.LLMRef)
		assert.Equal(t, "chain_memory", chain.MemoryRef)
		assert.Equal(t, "chain_prompt", chain.PromptRef)
	})

	t.Run("生成Agent配置", func(t *testing.T) {
		template := AgentTemplate{
			Type: "zero_shot_react",
			LLMTemplate: LLMTemplate{
				Type:  "openai",
				Model: "gpt-4",
			},
			MemoryType: "simple",
			MaxSteps:   3,
		}

		filename := "test_agent.json"
		err := generator.GenerateAgentConfig(template, filename)
		require.NoError(t, err)

		// 验证文件是否生成
		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)

		// 验证内容
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var config Config
		err = json.Unmarshal(data, &config)
		require.NoError(t, err)

		// 验证Agent
		assert.Len(t, config.Agents, 1)
		agent := config.Agents["main_agent"]
		assert.Equal(t, "zero_shot_react", agent.Type)
		assert.Equal(t, "agent_llm", agent.LLMRef)
		assert.Equal(t, "agent_memory", agent.MemoryRef)
		assert.Equal(t, 3, *agent.MaxSteps)
	})

	t.Run("生成Executor配置", func(t *testing.T) {
		template := ExecutorTemplate{
			AgentTemplate: AgentTemplate{
				Type: "conversational_react",
				LLMTemplate: LLMTemplate{
					Type:  "qwen",
					Model: "qwen-plus",
				},
				MaxSteps: 5,
			},
			MaxIterations:           8,
			ReturnIntermediateSteps: true,
		}

		filename := "test_executor.json"
		err := generator.GenerateExecutorConfig(template, filename)
		require.NoError(t, err)

		// 验证文件是否生成
		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)

		// 验证内容
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var config ExecutorUsageConfig
		err = json.Unmarshal(data, &config)
		require.NoError(t, err)

		// 验证Executor配置
		assert.NotNil(t, config.Agent)
		assert.Equal(t, "conversational_react", config.Agent.Type)
		assert.Equal(t, 8, *config.MaxIterations)
		assert.True(t, *config.ReturnIntermediateSteps)

		// 验证Agent配置
		assert.NotNil(t, config.Agent.LLM)
		assert.Equal(t, "qwen", config.Agent.LLM.Type)
		assert.Equal(t, "qwen-plus", config.Agent.LLM.Model)
		assert.Equal(t, 5, *config.Agent.MaxSteps)
	})
}

func TestConfigGeneratorPresets(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	generator := NewConfigGenerator(tempDir)

	t.Run("生成DeepSeek聊天配置", func(t *testing.T) {
		filename := "deepseek_chat.json"
		err := generator.GenerateDeepSeekChatConfig(filename)
		require.NoError(t, err)

		// 验证文件生成
		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)

		// 验证内容
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var config Config
		err = json.Unmarshal(data, &config)
		require.NoError(t, err)

		// 验证配置内容
		assert.Len(t, config.LLMs, 1)
		assert.Len(t, config.Memories, 1)
		assert.Len(t, config.Chains, 1)

		llm := config.LLMs["chain_llm"]
		assert.Equal(t, "deepseek", llm.Type)
		assert.Equal(t, "deepseek-chat", llm.Model)
	})

	t.Run("生成Kimi聊天配置", func(t *testing.T) {
		filename := "kimi_chat.json"
		err := generator.GenerateKimiChatConfig(filename)
		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)
	})

	t.Run("生成OpenAI聊天配置", func(t *testing.T) {
		filename := "openai_chat.json"
		err := generator.GenerateOpenAIChatConfig(filename)
		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)
	})

	t.Run("生成ReAct智能体配置", func(t *testing.T) {
		filename := "react_agent.json"
		err := generator.GenerateReactAgentConfig("deepseek", "deepseek-chat", filename)
		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)
	})

	t.Run("生成DeepSeek执行器配置", func(t *testing.T) {
		filename := "deepseek_executor.json"
		err := generator.GenerateExecutorWithDeepSeek(filename)
		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)
	})
}

func TestQuickGenerateMethods(t *testing.T) {
	// 创建临时目录并设置为默认生成器的输出目录
	tempDir := t.TempDir()
	oldGenerator := defaultGenerator
	defaultGenerator = NewConfigGenerator(tempDir)
	defer func() {
		defaultGenerator = oldGenerator
	}()

	t.Run("快速生成LLM配置", func(t *testing.T) {
		filename := "quick_llm.json"
		err := QuickGenerateLLM("anthropic", "claude-3-sonnet-20240229", filename)
		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)

		// 验证内容
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var config Config
		err = json.Unmarshal(data, &config)
		require.NoError(t, err)

		llm := config.LLMs["main_llm"]
		assert.Equal(t, "anthropic", llm.Type)
		assert.Equal(t, "claude-3-sonnet-20240229", llm.Model)
	})

	t.Run("快速生成Chain配置", func(t *testing.T) {
		filename := "quick_chain.json"
		err := QuickGenerateChain("llm", "ollama", "llama2", filename)
		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)
	})

	t.Run("快速生成Agent配置", func(t *testing.T) {
		filename := "quick_agent.json"
		err := QuickGenerateAgent("zero_shot_react", "deepseek", "deepseek-chat", filename)
		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)
	})

	t.Run("快速生成Executor配置", func(t *testing.T) {
		filename := "quick_executor.json"
		err := QuickGenerateExecutor("conversational_react", "kimi", "moonshot-v1-8k", filename)
		require.NoError(t, err)

		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)
	})
}

func TestConfigGeneratorValidation(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewConfigGenerator(tempDir)

	t.Run("LLM类型为空应该失败", func(t *testing.T) {
		template := LLMTemplate{
			Model: "test-model",
		}
		err := generator.GenerateLLMConfig(template, "invalid.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LLM type is required")
	})

	t.Run("模型名称为空应该失败", func(t *testing.T) {
		template := LLMTemplate{
			Type: "openai",
		}
		err := generator.GenerateLLMConfig(template, "invalid.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model is required")
	})

	t.Run("Chain类型为空应该失败", func(t *testing.T) {
		template := ChainTemplate{
			LLMTemplate: LLMTemplate{
				Type:  "openai",
				Model: "gpt-4",
			},
		}
		err := generator.GenerateChainConfig(template, "invalid.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Chain type is required")
	})
}

func TestGetDefaultAPIKeyEnv(t *testing.T) {
	generator := NewConfigGenerator(".")

	tests := []struct {
		llmType  string
		expected string
	}{
		{"openai", "${OPENAI_API_KEY}"},
		{"deepseek", "${DEEPSEEK_API_KEY}"},
		{"kimi", "${KIMI_API_KEY}"},
		{"qwen", "${QWEN_API_KEY}"},
		{"anthropic", "${ANTHROPIC_API_KEY}"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.llmType, func(t *testing.T) {
			result := generator.getDefaultAPIKeyEnv(tt.llmType)
			assert.Equal(t, tt.expected, result)
		})
	}
}