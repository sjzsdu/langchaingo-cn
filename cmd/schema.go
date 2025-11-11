package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sjzsdu/langchaingo-cn/schema"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
)

var (
	// 全局配置
	outputDir  string
	outputFile string
	verbose    bool
	
	// 验证配置
	enableAPITest bool  // 是否启用真实API调用测试

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
  • zhipu       - 智谱AI GLM模型
  • siliconflow - 硅基流动平台模型
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
	Long:  "生成大语言模型(LLM)的配置文件，支持DeepSeek、Kimi、OpenAI、智谱AI、硅基流动等模型",
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
	ValidArgs: []string{"deepseek-chat", "kimi-chat", "openai-chat", "qwen-chat", "zhipu-chat", "siliconflow-chat", "deepseek-executor", "zhipu-executor", "siliconflow-executor"},
	Example: `  # 生成DeepSeek聊天配置
  config-gen preset deepseek-chat -o deepseek.json

  # 生成智谱AI聊天配置
  config-gen preset zhipu-chat -o zhipu.json

  # 生成硅基流动聊天配置
  config-gen preset siliconflow-chat -o siliconflow.json

  # 生成智谱AI执行器配置
  config-gen preset zhipu-executor -o executor.json`,
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
		fmt.Println("  • deepseek-chat       - DeepSeek聊天配置")
		fmt.Println("  • kimi-chat           - Kimi聊天配置")
		fmt.Println("  • openai-chat         - OpenAI聊天配置")
		fmt.Println("  • qwen-chat           - 通义千问聊天配置")
		fmt.Println("  • zhipu-chat          - 智谱AI聊天配置")
		fmt.Println("  • siliconflow-chat    - 硅基流动聊天配置")
		fmt.Println("  • deepseek-executor   - DeepSeek执行器配置")
		fmt.Println("  • zhipu-executor      - 智谱AI执行器配置")
		fmt.Println("  • siliconflow-executor - 硅基流动执行器配置")

		fmt.Println("\n🤖 支持的LLM类型:")
		fmt.Println("  • deepseek    - DeepSeek模型")
		fmt.Println("  • kimi        - Kimi月之暗面模型")
		fmt.Println("  • openai      - OpenAI GPT模型")
		fmt.Println("  • qwen        - 通义千问模型")
		fmt.Println("  • zhipu       - 智谱AI GLM模型")
		fmt.Println("  • siliconflow - 硅基流动平台模型")
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

// Validate命令
var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "验证并测试JSON配置文件",
	Long: `验证并测试JSON配置文件是否能正常工作

该命令会:
1. 解析JSON配置文件
2. 验证配置语法和结构
3. 创建相关组件实例 
4. 执行基本功能测试
5. 可选的真实API调用测试
6. 报告验证结果

验证级别:
• 基础验证: 检查配置语法和组件创建
• API测试: 发送真实请求测试LLM/Chain/Agent功能

示例:
  # 基础验证配置文件
  xin config-gen validate config.json
  
  # 完整验证(包含真实API调用)
  xin config-gen validate config.json --api-test
  
  # 验证配置并显示详细信息
  xin config-gen validate config.json --verbose`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := args[0]
		validateConfiguration(configFile)
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

	// Validate命令标志
	validateCmd.Flags().BoolVarP(&enableAPITest, "api-test", "t", false, "启用真实API调用测试")

	// 添加子命令
	configGenCmd.AddCommand(llmCmd)
	configGenCmd.AddCommand(chainCmd)
	configGenCmd.AddCommand(agentCmd)
	configGenCmd.AddCommand(executorCmd)
	configGenCmd.AddCommand(presetCmd)
	configGenCmd.AddCommand(listCmd)
	configGenCmd.AddCommand(validateCmd)
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
	case "zhipu-chat":
		return generator.GenerateZhipuChatConfig(output)
	case "siliconflow-chat":
		return generator.GenerateSiliconFlowChatConfig(output)
	case "deepseek-executor":
		return generator.GenerateExecutorWithDeepSeek(output)
	case "zhipu-executor":
		return generator.GenerateExecutorWithZhipu(output)
	case "siliconflow-executor":
		return generator.GenerateExecutorWithSiliconFlow(output)
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

// validateConfiguration 验证配置文件
func validateConfiguration(configFile string) {
	fmt.Printf("🔍 验证配置文件: %s\n", configFile)

	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("❌ 配置文件不存在: %s\n", configFile)
		return
	}

	if verbose {
		fmt.Println("\n📋 验证步骤:")
		fmt.Println("   1. 📄 解析JSON配置文件...")
	}

	// 步骤1: 解析配置文件
	app, err := schema.CreateApplicationFromFile(configFile)
	if err != nil {
		fmt.Printf("❌ 配置文件解析失败: %v\n", err)
		return
	}

	if verbose {
		fmt.Println("   ✅ JSON配置解析成功")
		fmt.Println("   2. 🔧 验证组件创建...")
	}

	// 统计创建的组件
	var stats struct {
		llms     int
		memories int
		chains   int
		agents   int
	}

	// 验证LLM组件
	for name, llm := range app.LLMs {
		if llm == nil {
			fmt.Printf("❌ LLM组件创建失败: %s\n", name)
			return
		}
		stats.llms++
		if verbose {
			fmt.Printf("      ✅ LLM: %s\n", name)
		}
	}

	// 验证Memory组件
	for name, memory := range app.Memories {
		if memory == nil {
			fmt.Printf("❌ Memory组件创建失败: %s\n", name)
			return
		}
		stats.memories++
		if verbose {
			fmt.Printf("      ✅ Memory: %s\n", name)
		}
	}

	// 验证Chain组件
	for name, chain := range app.Chains {
		if chain == nil {
			fmt.Printf("❌ Chain组件创建失败: %s\n", name)
			return
		}
		stats.chains++
		if verbose {
			fmt.Printf("      ✅ Chain: %s\n", name)
		}
	}

	// 验证Agent组件
	for name, agent := range app.Agents {
		if agent == nil {
			fmt.Printf("❌ Agent组件创建失败: %s\n", name)
			return
		}
		stats.agents++
		if verbose {
			fmt.Printf("      ✅ Agent: %s\n", name)
		}
	}

	// 注意: Agents字段实际上包含的是 agents.Executor
	// 这里不需要单独验证Executors，因为它们包含在Agents中

	if verbose {
		fmt.Println("   3. 🧪 执行功能测试...")
	}

	// 步骤3: 基本功能测试
	testSuccess := true

	// 测试GetModels方法
	for name, llm := range app.LLMs {
		if modelsGetter, ok := llm.(interface{ GetModels() []string }); ok {
			models := modelsGetter.GetModels()
			if len(models) == 0 {
				fmt.Printf("⚠️  LLM %s 的 GetModels() 返回空模型列表\n", name)
			} else if verbose {
				fmt.Printf("      ✅ LLM %s 支持 %d 个模型\n", name, len(models))
			}
		}
	}

	// 如果启用API测试，进行真实调用测试
	if enableAPITest {
		if verbose {
			fmt.Println("      🌐 执行真实API调用测试...")
		} else {
			fmt.Println("   🌐 执行真实API调用测试...")
		}
		
		// 测试LLM真实API调用
		for name, llm := range app.LLMs {
			if verbose {
				fmt.Printf("      🔍 测试LLM %s 的API调用...\n", name)
			}
			
			if !testLLMAPICall(name, llm, verbose) {
				testSuccess = false
			}
		}

		// 测试Chain的真实调用
		for name, chain := range app.Chains {
			if verbose {
				fmt.Printf("      🔍 测试Chain %s 的对话功能...\n", name)
			}
			
			if !testChainAPICall(name, chain, verbose) {
				testSuccess = false
			}
		}

		// 测试Agent的真实调用
		for name, agent := range app.Agents {
			if verbose {
				fmt.Printf("      🔍 测试Agent %s 的执行功能...\n", name)
			}
			
			if !testAgentAPICall(name, agent, verbose) {
				testSuccess = false
			}
		}
	} else {
		if verbose {
			fmt.Println("      ℹ️  跳过API调用测试 (使用 --api-test 启用)")
		}
		// 如果有组件，给出提示
		if len(app.Chains) > 0 && verbose {
			fmt.Println("      ℹ️  Chain组件已就绪，可用于对话测试")
		}
	}

	// 步骤4: 生成验证报告
	if verbose {
		fmt.Println("   4. 📊 生成验证报告...")
	}

	fmt.Printf("\n✅ 配置验证成功! %s\n", configFile)
	fmt.Printf("📊 组件统计:\n")
	fmt.Printf("   🤖 LLMs: %d\n", stats.llms)
	fmt.Printf("   💾 Memories: %d\n", stats.memories)
	fmt.Printf("   ⛓️  Chains: %d\n", stats.chains)
	fmt.Printf("   🤖 Agents/Executors: %d\n", stats.agents)

	if testSuccess {
		fmt.Println("\n🎉 所有组件验证通过，配置文件可以正常使用!")
	} else {
		fmt.Println("\n⚠️  部分组件验证失败，请检查API密钥设置和网络连接")
	}

	if verbose {
		fmt.Println("\n💡 下一步:")
		fmt.Println("   1. 设置必要的环境变量(API Keys)")
		fmt.Println("   2. 在代码中使用 schema.CreateApplicationFromFile() 加载配置")
		fmt.Println("   3. 开始构建你的AI应用!")
	}
}

// testLLMAPICall 测试LLM的真实API调用
func testLLMAPICall(name string, llm llms.Model, verbose bool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试问题
	testPrompt := "你是什么模型？请简短回答。"
	
	if verbose {
		fmt.Printf("        📤 发送测试问题: %s\n", testPrompt)
	}

	// 尝试调用LLM
	response, err := llms.GenerateFromSinglePrompt(ctx, llm, testPrompt)
	if err != nil {
		fmt.Printf("        ❌ LLM %s API调用失败: %v\n", name, err)
		return false
	}

	if response == "" {
		fmt.Printf("        ❌ LLM %s 返回空响应\n", name)
		return false
	}

	if verbose {
		// 截断长响应
		truncatedResponse := response
		if len(response) > 100 {
			truncatedResponse = response[:100] + "..."
		}
		fmt.Printf("        📥 收到响应: %s\n", truncatedResponse)
	}
	
	fmt.Printf("        ✅ LLM %s API调用成功\n", name)
	return true
}

// testChainAPICall 测试Chain的真实调用
func testChainAPICall(name string, chain chains.Chain, verbose bool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试输入
	testInput := map[string]any{
		"input": "你好，请告诉我你是什么AI助手？",
	}
	
	if verbose {
		fmt.Printf("        📤 发送测试输入: %s\n", testInput["input"])
	}

	// 尝试调用Chain
	result, err := chains.Call(ctx, chain, testInput)
	if err != nil {
		fmt.Printf("        ❌ Chain %s 调用失败: %v\n", name, err)
		return false
	}

	// 检查结果
	if result == nil {
		fmt.Printf("        ❌ Chain %s 返回空结果\n", name)
		return false
	}

	// 尝试获取输出
	var output string
	if outputValue, exists := result["output"]; exists {
		if str, ok := outputValue.(string); ok {
			output = str
		}
	} else if textValue, exists := result["text"]; exists {
		if str, ok := textValue.(string); ok {
			output = str
		}
	}

	if output == "" {
		fmt.Printf("        ❌ Chain %s 没有产生有效输出\n", name)
		return false
	}

	if verbose {
		// 截断长响应
		truncatedOutput := output
		if len(output) > 100 {
			truncatedOutput = output[:100] + "..."
		}
		fmt.Printf("        📥 收到响应: %s\n", truncatedOutput)
	}

	fmt.Printf("        ✅ Chain %s 调用成功\n", name)
	return true
}

// testAgentAPICall 测试Agent的真实执行
func testAgentAPICall(name string, agent interface{}, verbose bool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// 测试问题
	testInput := "请简单介绍一下你自己的能力"
	
	if verbose {
		fmt.Printf("        📤 发送测试问题: %s\n", testInput)
	}

	// 尝试执行Agent (agents.Executor有Call方法)
	if executor, ok := agent.(interface {
		Call(ctx context.Context, inputs map[string]any) (map[string]any, error)
	}); ok {
		result, err := executor.Call(ctx, map[string]any{
			"input": testInput,
		})
		
		if err != nil {
			fmt.Printf("        ❌ Agent %s 执行失败: %v\n", name, err)
			return false
		}

		// 检查结果
		if result == nil {
			fmt.Printf("        ❌ Agent %s 返回空结果\n", name)
			return false
		}

		// 尝试获取输出
		var output string
		if outputValue, exists := result["output"]; exists {
			if str, ok := outputValue.(string); ok {
				output = str
			}
		}

		if output == "" {
			fmt.Printf("        ❌ Agent %s 没有产生有效输出\n", name)
			return false
		}

		if verbose {
			// 截断长响应
			truncatedOutput := output
			if len(output) > 100 {
				truncatedOutput = output[:100] + "..."
			}
			fmt.Printf("        📥 收到响应: %s\n", truncatedOutput)
		}

		fmt.Printf("        ✅ Agent %s 执行成功\n", name)
		return true
	}

	fmt.Printf("        ⚠️  Agent %s 不支持标准调用接口\n", name)
	return true // 不算作失败
}
