package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/sjzsdu/langchaingo-cn/schema"
	"github.com/spf13/cobra"
)

var (
	// 全局配置
	outputDir  string
	outputFile string
	verbose    bool

	// LLM配置
	llmType     string
	model       string
	temperature float64
	maxTokens   int
	baseURL     string

	// Chain配置
	chainType      string
	memoryType     string
	promptTemplate string

	// Agent配置
	agentType string
	maxSteps  int

	// Executor配置
	maxIterations           int
	returnIntermediateSteps bool
)

var configGenCmd = &cobra.Command{
	Use:   "config-gen",
	Short: "🚀 LangChainGo-CN 配置文件生成器",
	Long: `🚀 LangChainGo-CN 配置文件生成器

一个强大的工具，帮助您快速生成LangChain组件配置文件。
支持LLM、Chain、Agent、Executor等组件的配置生成。

支持的LLM类型:
  • deepseek    - DeepSeek模型
  • kimi        - Kimi月之暗面模型
  • openai      - OpenAI GPT模型
  • qwen        - 通义千问模型
  • anthropic   - Anthropic Claude模型
  • ollama      - 本地Ollama模型`,
	Example: `  # 生成DeepSeek聊天配置
  langchaingo-cn config-gen preset deepseek-chat -o deepseek.json

  # 生成自定义LLM配置
  langchaingo-cn config-gen llm --llm deepseek --model deepseek-chat -o my_llm.json

  # 生成Chain配置
  langchaingo-cn config-gen chain --llm kimi --model moonshot-v1-8k --memory conversation_buffer

  # 生成Agent配置
  langchaingo-cn config-gen agent --llm openai --model gpt-4 --agent-type zero_shot_react`,
}

// LLM命令
var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "生成LLM配置文件",
	Long:  "生成大语言模型(LLM)的配置文件，支持DeepSeek、Kimi、OpenAI等模型",
	Example: `  # 生成DeepSeek配置
  config-gen llm --llm deepseek --model deepseek-chat

  # 生成带参数的OpenAI配置
  config-gen llm --llm openai --model gpt-4 --temperature 0.7 --max-tokens 2048`,
	Run: func(cmd *cobra.Command, args []string) {
		if llmType == "" {
			log.Fatal("❌ 请指定LLM类型 (--llm)")
		}
		if model == "" {
			log.Fatal("❌ 请指定模型名称 (--model)")
		}

		generator := schema.NewConfigGenerator(outputDir)
		template := schema.LLMTemplate{
			Type:  llmType,
			Model: model,
		}

		if temperature > 0 {
			template.Temperature = temperature
		}
		if maxTokens > 0 {
			template.MaxTokens = maxTokens
		}
		if baseURL != "" {
			template.BaseURL = baseURL
		}

		if err := generator.GenerateLLMConfig(template, outputFile); err != nil {
			log.Fatal("❌ 生成LLM配置失败:", err)
		}

		printSuccess("LLM配置")
	},
}

// Chain命令
var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "生成Chain配置文件",
	Long:  "生成链(Chain)的配置文件，支持对话链、LLM链等类型",
	Example: `  # 生成对话链配置
  config-gen chain --llm deepseek --model deepseek-chat --memory conversation_buffer

  # 生成带自定义提示的链配置
  config-gen chain --llm kimi --model moonshot-v1-8k --prompt "你是专业助手：{{.input}}"`,
	Run: func(cmd *cobra.Command, args []string) {
		if llmType == "" {
			log.Fatal("❌ 请指定LLM类型 (--llm)")
		}
		if model == "" {
			log.Fatal("❌ 请指定模型名称 (--model)")
		}

		generator := schema.NewConfigGenerator(outputDir)
		template := schema.ChainTemplate{
			Type: getChainType(),
			LLMTemplate: schema.LLMTemplate{
				Type:        llmType,
				Model:       model,
				Temperature: temperature,
			},
		}

		if maxTokens > 0 {
			template.LLMTemplate.MaxTokens = maxTokens
		}
		if baseURL != "" {
			template.LLMTemplate.BaseURL = baseURL
		}
		if memoryType != "" {
			template.MemoryType = memoryType
		}
		if promptTemplate != "" {
			template.PromptTemplate = promptTemplate
			template.InputVariables = []string{"input"}
		}

		if err := generator.GenerateChainConfig(template, outputFile); err != nil {
			log.Fatal("❌ 生成Chain配置失败:", err)
		}

		printSuccess("Chain配置")
	},
}

// Agent命令
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "生成Agent配置文件",
	Long:  "生成智能体(Agent)的配置文件，支持零样本ReAct、对话ReAct等类型",
	Example: `  # 生成零样本ReAct智能体
  config-gen agent --llm openai --model gpt-4 --agent-type zero_shot_react

  # 生成对话智能体
  config-gen agent --llm deepseek --model deepseek-chat --agent-type conversational_react --max-steps 5`,
	Run: func(cmd *cobra.Command, args []string) {
		if llmType == "" {
			log.Fatal("❌ 请指定LLM类型 (--llm)")
		}
		if model == "" {
			log.Fatal("❌ 请指定模型名称 (--model)")
		}

		generator := schema.NewConfigGenerator(outputDir)
		template := schema.AgentTemplate{
			Type: getAgentType(),
			LLMTemplate: schema.LLMTemplate{
				Type:        llmType,
				Model:       model,
				Temperature: temperature,
			},
			MemoryType: memoryType,
			MaxSteps:   maxSteps,
		}

		if maxTokens > 0 {
			template.LLMTemplate.MaxTokens = maxTokens
		}
		if baseURL != "" {
			template.LLMTemplate.BaseURL = baseURL
		}

		if err := generator.GenerateAgentConfig(template, outputFile); err != nil {
			log.Fatal("❌ 生成Agent配置失败:", err)
		}

		printSuccess("Agent配置")
	},
}

// Executor命令
var executorCmd = &cobra.Command{
	Use:   "executor",
	Short: "生成Executor配置文件",
	Long:  "生成执行器(Executor)的配置文件，使用新的usage风格配置",
	Example: `  # 生成基本执行器配置
  config-gen executor --llm deepseek --model deepseek-chat

  # 生成带详细参数的执行器配置
  config-gen executor --llm kimi --model moonshot-v1-8k --max-iterations 10 --return-steps`,
	Run: func(cmd *cobra.Command, args []string) {
		if llmType == "" {
			log.Fatal("❌ 请指定LLM类型 (--llm)")
		}
		if model == "" {
			log.Fatal("❌ 请指定模型名称 (--model)")
		}

		generator := schema.NewConfigGenerator(outputDir)
		template := schema.ExecutorTemplate{
			AgentTemplate: schema.AgentTemplate{
				Type: getAgentType(),
				LLMTemplate: schema.LLMTemplate{
					Type:        llmType,
					Model:       model,
					Temperature: temperature,
				},
				MemoryType: memoryType,
				MaxSteps:   maxSteps,
			},
			MaxIterations:           maxIterations,
			ReturnIntermediateSteps: returnIntermediateSteps,
		}

		if maxTokens > 0 {
			template.AgentTemplate.LLMTemplate.MaxTokens = maxTokens
		}
		if baseURL != "" {
			template.AgentTemplate.LLMTemplate.BaseURL = baseURL
		}

		if err := generator.GenerateExecutorConfig(template, outputFile); err != nil {
			log.Fatal("❌ 生成Executor配置失败:", err)
		}

		printSuccess("Executor配置")
	},
}

// Preset命令
var presetCmd = &cobra.Command{
	Use:       "preset [配置名称]",
	Short:     "生成预设配置文件",
	Long:      "使用预定义的配置模板快速生成常用配置文件",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"deepseek-chat", "kimi-chat", "openai-chat", "qwen-chat", "deepseek-executor"},
	Example: `  # 生成DeepSeek聊天配置
  config-gen preset deepseek-chat -o deepseek.json

  # 生成Kimi聊天配置
  config-gen preset kimi-chat -o kimi.json

  # 生成DeepSeek执行器配置
  config-gen preset deepseek-executor -o executor.json`,
	Run: func(cmd *cobra.Command, args []string) {
		preset := args[0]
		generator := schema.NewConfigGenerator(outputDir)

		if err := generatePreset(generator, preset, outputFile); err != nil {
			log.Fatal("❌ 生成预设配置失败:", err)
		}

		printSuccess(fmt.Sprintf("预设配置 [%s]", preset))
	},
}

// List命令
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用的预设配置和支持的类型",
	Long:  "显示所有可用的预设配置、支持的LLM类型、Chain类型等信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📋 可用的预设配置:")
		fmt.Println("  • deepseek-chat     - DeepSeek聊天配置")
		fmt.Println("  • kimi-chat         - Kimi聊天配置")
		fmt.Println("  • openai-chat       - OpenAI聊天配置")
		fmt.Println("  • qwen-chat         - 通义千问聊天配置")
		fmt.Println("  • deepseek-executor - DeepSeek执行器配置")

		fmt.Println("\n🤖 支持的LLM类型:")
		fmt.Println("  • deepseek    - DeepSeek模型")
		fmt.Println("  • kimi        - Kimi月之暗面模型")
		fmt.Println("  • openai      - OpenAI GPT模型")
		fmt.Println("  • qwen        - 通义千问模型")
		fmt.Println("  • anthropic   - Anthropic Claude模型")
		fmt.Println("  • ollama      - 本地Ollama模型")

		fmt.Println("\n⛓️  支持的Chain类型:")
		fmt.Println("  • conversation - 对话链")
		fmt.Println("  • llm          - LLM链")
		fmt.Println("  • sequential   - 顺序链")

		fmt.Println("\n🤖 支持的Agent类型:")
		fmt.Println("  • zero_shot_react      - 零样本ReAct智能体")
		fmt.Println("  • conversational_react - 对话ReAct智能体")

		fmt.Println("\n💾 支持的Memory类型:")
		fmt.Println("  • conversation_buffer - 会话缓冲记忆")
		fmt.Println("  • simple              - 简单记忆")
	},
}

func init() {
	// 全局标志
	configGenCmd.PersistentFlags().StringVarP(&outputDir, "dir", "d", ".", "输出目录")
	configGenCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "config.json", "输出文件名")
	configGenCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")

	// LLM命令标志
	llmCmd.Flags().StringVar(&llmType, "llm", "", "LLM类型 (deepseek|kimi|openai|qwen|anthropic|ollama) [必需]")
	llmCmd.Flags().StringVar(&model, "model", "", "模型名称 [必需]")
	llmCmd.Flags().Float64Var(&temperature, "temperature", 0, "温度参数 (0.0-2.0)")
	llmCmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "最大token数")
	llmCmd.Flags().StringVar(&baseURL, "base-url", "", "自定义API基础URL")
	llmCmd.MarkFlagRequired("llm")
	llmCmd.MarkFlagRequired("model")

	// Chain命令标志
	chainCmd.Flags().StringVar(&llmType, "llm", "", "LLM类型 [必需]")
	chainCmd.Flags().StringVar(&model, "model", "", "模型名称 [必需]")
	chainCmd.Flags().StringVar(&chainType, "chain-type", "conversation", "Chain类型 (conversation|llm|sequential)")
	chainCmd.Flags().StringVar(&memoryType, "memory", "conversation_buffer", "Memory类型 (conversation_buffer|simple)")
	chainCmd.Flags().StringVar(&promptTemplate, "prompt", "", "自定义提示模板")
	chainCmd.Flags().Float64Var(&temperature, "temperature", 0.7, "温度参数")
	chainCmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "最大token数")
	chainCmd.Flags().StringVar(&baseURL, "base-url", "", "自定义API基础URL")
	chainCmd.MarkFlagRequired("llm")
	chainCmd.MarkFlagRequired("model")

	// Agent命令标志
	agentCmd.Flags().StringVar(&llmType, "llm", "", "LLM类型 [必需]")
	agentCmd.Flags().StringVar(&model, "model", "", "模型名称 [必需]")
	agentCmd.Flags().StringVar(&agentType, "agent-type", "zero_shot_react", "Agent类型 (zero_shot_react|conversational_react)")
	agentCmd.Flags().StringVar(&memoryType, "memory", "conversation_buffer", "Memory类型")
	agentCmd.Flags().IntVar(&maxSteps, "max-steps", 5, "最大步数")
	agentCmd.Flags().Float64Var(&temperature, "temperature", 0.3, "温度参数")
	agentCmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "最大token数")
	agentCmd.Flags().StringVar(&baseURL, "base-url", "", "自定义API基础URL")
	agentCmd.MarkFlagRequired("llm")
	agentCmd.MarkFlagRequired("model")

	// Executor命令标志
	executorCmd.Flags().StringVar(&llmType, "llm", "", "LLM类型 [必需]")
	executorCmd.Flags().StringVar(&model, "model", "", "模型名称 [必需]")
	executorCmd.Flags().StringVar(&agentType, "agent-type", "zero_shot_react", "Agent类型")
	executorCmd.Flags().StringVar(&memoryType, "memory", "conversation_buffer", "Memory类型")
	executorCmd.Flags().IntVar(&maxSteps, "max-steps", 5, "最大步数")
	executorCmd.Flags().IntVar(&maxIterations, "max-iterations", 10, "最大迭代次数")
	executorCmd.Flags().BoolVar(&returnIntermediateSteps, "return-steps", false, "返回中间步骤")
	executorCmd.Flags().Float64Var(&temperature, "temperature", 0.7, "温度参数")
	executorCmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "最大token数")
	executorCmd.Flags().StringVar(&baseURL, "base-url", "", "自定义API基础URL")
	executorCmd.MarkFlagRequired("llm")
	executorCmd.MarkFlagRequired("model")

	// 添加子命令
	configGenCmd.AddCommand(llmCmd)
	configGenCmd.AddCommand(chainCmd)
	configGenCmd.AddCommand(agentCmd)
	configGenCmd.AddCommand(executorCmd)
	configGenCmd.AddCommand(presetCmd)
	configGenCmd.AddCommand(listCmd)
}

// 辅助函数
func getChainType() string {
	if chainType == "" {
		return "conversation"
	}
	return chainType
}

func getAgentType() string {
	if agentType == "" {
		return "zero_shot_react"
	}
	return agentType
}

func generatePreset(generator *schema.ConfigGenerator, preset, output string) error {
	switch strings.ToLower(preset) {
	case "deepseek-chat":
		return generator.GenerateDeepSeekChatConfig(output)
	case "kimi-chat":
		return generator.GenerateKimiChatConfig(output)
	case "openai-chat":
		return generator.GenerateOpenAIChatConfig(output)
	case "qwen-chat":
		return generator.GenerateQwenChatConfig(output)
	case "deepseek-executor":
		return generator.GenerateExecutorWithDeepSeek(output)
	default:
		return fmt.Errorf("不支持的预设配置: %s", preset)
	}
}

func printSuccess(configType string) {
	if verbose {
		fmt.Printf("\n✨ %s生成成功!\n", configType)
		fmt.Printf("📁 文件位置: %s/%s\n", outputDir, outputFile)
		fmt.Println("\n💡 使用指南:")
		fmt.Println("   1. 🔑 设置相应的环境变量 (如 DEEPSEEK_API_KEY)")
		fmt.Println("   2. 📝 使用 schema.CreateApplicationFromFile() 加载配置")
		fmt.Println("   3. 🚀 开始使用您的AI应用!")
	} else {
		fmt.Printf("✅ %s已生成: %s/%s\n", configType, outputDir, outputFile)
	}
}
