package llms_test

import (
	"testing"

	llmscn "github.com/sjzsdu/langchaingo-cn/llms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLLM(t *testing.T) {
	// 测试创建DeepSeek LLM
	t.Run("CreateDeepSeekLLM", func(t *testing.T) {
		// 缺少必要参数的情况
		_, err := llmscn.CreateLLM(llmscn.DeepSeekLLM, map[string]interface{}{})
		require.Error(t, err)
		assert.ErrorIs(t, err, llmscn.ErrMissingRequiredParam)

		// 提供必要参数的情况
		llm, err := llmscn.CreateLLM(llmscn.DeepSeekLLM, map[string]interface{}{
			"api_key": "test-api-key",
			"model":   "deepseek-chat",
		})
		require.NoError(t, err)
		require.NotNil(t, llm)
	})

	// 测试创建Kimi LLM
	t.Run("CreateKimiLLM", func(t *testing.T) {
		// 缺少必要参数的情况
		_, err := llmscn.CreateLLM(llmscn.KimiLLM, map[string]interface{}{})
		require.Error(t, err)
		assert.ErrorIs(t, err, llmscn.ErrMissingRequiredParam)

		// 提供必要参数的情况
		llm, err := llmscn.CreateLLM(llmscn.KimiLLM, map[string]interface{}{
			"api_key":      "test-api-key",
			"model":        "moonshot-v1-8k",
			"temperature":  0.7,
			"top_p":        0.9,
			"max_tokens":   1000,
		})
		require.NoError(t, err)
		require.NotNil(t, llm)
	})

	// 测试创建Qwen LLM
	t.Run("CreateQwenLLM", func(t *testing.T) {
		// 缺少必要参数的情况
		_, err := llmscn.CreateLLM(llmscn.QwenLLM, map[string]interface{}{})
		require.Error(t, err)
		assert.ErrorIs(t, err, llmscn.ErrMissingRequiredParam)

		// 提供必要参数的情况
		llm, err := llmscn.CreateLLM(llmscn.QwenLLM, map[string]interface{}{
			"api_key":              "test-api-key",
			"model":                "qwen-turbo",
			"temperature":          0.8,
			"top_p":                0.95,
			"top_k":                50,
			"max_tokens":           2000,
			"use_openai_compatible": true,
		})
		require.NoError(t, err)
		require.NotNil(t, llm)
	})

	// 测试不支持的LLM类型
	t.Run("UnsupportedLLMType", func(t *testing.T) {
		_, err := llmscn.CreateLLM("unsupported", map[string]interface{}{
			"api_key": "test-api-key",
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, llmscn.ErrUnsupportedLLMType)
	})
}