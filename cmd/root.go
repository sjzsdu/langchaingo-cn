package cmd

import (
	"fmt"
	"os"

	"github.com/sjzsdu/langchaingo-cn/share"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   share.BUILDNAME,
	Short: "🚀 LangChainGo-CN - 中文化的 LangChain Go 版本",
	Long: `🚀 LangChainGo-CN - 中文化的 LangChain Go 版本

一个功能强大的 Go 语言 LangChain 扩展，专为中文用户优化。
提供完整的配置管理、模型集成和工作流编排能力。

主要功能:
  • 🤖 支持多种 LLM 模型 (DeepSeek、Kimi、OpenAI、Qwen 等)
  • ⛓️  灵活的链式处理能力
  • 🧠 智能体 (Agent) 系统
  • 📊 图形化工作流 (Graph)
  • 💾 多种记忆类型支持
  • 🔧 配置文件生成工具`,
	Version: share.VERSION,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Add subcommands
	rootCmd.AddCommand(configGenCmd)
}
