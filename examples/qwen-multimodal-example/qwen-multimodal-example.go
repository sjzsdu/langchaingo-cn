// 通义千问多模态示例 - 展示如何使用通义千问多模态模型处理图像和文本
//
// 本示例展示了以下功能：
// 1. 基本多模态图片分析 - 使用远程图片URL进行单轮图像分析
// 2. 多轮多模态对话 - 在对话上下文中使用图片并进行多轮交互
// 3. 流式多模态输出 - 使用流式API实时接收模型响应
// 4. 本地图片处理 - 加载并分析本地图片文件
//
// 运行前准备：
// 1. 确保已设置QWEN_API_KEY环境变量
// 2. 确保example.jpg文件存在于当前目录
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sjzsdu/langchaingo-cn/llms/qwen"
	"github.com/tmc/langchaingo/llms"
)

// 定义常量
const (
	// 图片详细程度
	ImageDetailHigh = "high" // 高详细度，适用于需要细节分析的场景
	ImageDetailLow  = "low"  // 低详细度，适用于简单识别场景，可节省token

	// 远程图片URL - 使用固定的图片URL而非随机API
	// 注意：使用固定URL可以避免API错误
	NatureImageURL       = "https://images.unsplash.com/photo-1472214103451-9374bd1c798e?w=1024&h=768" // 固定自然风景图片
	ArchitectureImageURL = "https://images.unsplash.com/photo-1486325212027-8081e485255e?w=1024&h=768" // 固定城市建筑图片

	// 本地图片路径 - 相对于当前工作目录
	// 运行示例前，请确保此路径下存在有效的图像文件
	LocalImagePath = "./example.jpg"
)

// getImageAsBase64 从URL下载图片并转换为Base64格式的ImageURLContent
func getImageAsBase64(imageURL string, detail string) (llms.ImageURLContent, error) {
	// 尝试直接使用本地图片
	fmt.Println("改为使用本地图片...")
	
	// 获取本地图片的绝对路径
	localImagePath, err := filepath.Abs(LocalImagePath)
	if err != nil {
		return llms.ImageURLContent{}, fmt.Errorf("获取图片绝对路径失败: %v", err)
	}
	
	// 检查本地图片是否存在
	_, err = os.Stat(localImagePath)
	if os.IsNotExist(err) {
		return llms.ImageURLContent{}, fmt.Errorf("本地图片文件 %s 不存在", localImagePath)
	}
	
	// 读取图片文件
	imageData, err := os.ReadFile(localImagePath)
	if err != nil {
		return llms.ImageURLContent{}, fmt.Errorf("读取本地图片失败: %v", err)
	}
	
	// 将图片编码为base64
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)
	
	// 创建base64格式的图片内容
	base64URL := fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64)
	
	// 打印URL的前50个字符（用于调试）
	fmt.Printf("Base64 URL前缀: %s...（已截断）\n", base64URL[:50])
	
	return llms.ImageURLContent{
		URL:    base64URL,
		Detail: detail,
	}, nil
}

func main() {
	// 打印环境变量信息（不包含敏感信息）
	fmt.Println("=== 环境变量检查 ===")
	apiKey := os.Getenv("QWEN_API_KEY")
	if apiKey == "" {
		log.Fatalf("错误: 未设置QWEN_API_KEY环境变量\n")
	} else {
		// 只打印API密钥的前8位和后4位，中间用***替代
		maskedKey := apiKey
		if len(apiKey) > 12 {
			maskedKey = apiKey[:8] + "***" + apiKey[len(apiKey)-4:]
		}
		fmt.Printf("QWEN_API_KEY: %s (已设置)\n", maskedKey)
	}

	// 检查模型环境变量
	modelEnv := os.Getenv("QWEN_MODEL")
	modelToUse := qwen.ModelQWenVLPlus
	if modelEnv != "" {
		modelToUse = modelEnv
		fmt.Printf("使用环境变量指定的模型: %s\n", modelToUse)
	} else {
		fmt.Printf("使用默认模型: %s\n", modelToUse)
	}

	// 初始化通义千问多模态客户端
	fmt.Println("正在初始化通义千问客户端...")
	llm, err := qwen.New(
		qwen.WithModel(modelToUse), // 使用通义千问视觉语言模型
		qwen.WithOpenAICompatible(true), // 使用OpenAI兼容模式
		qwen.WithTemperature(0.7), // 设置温度
		qwen.WithMaxTokens(1000), // 设置最大token数
	)
	if err != nil {
		log.Fatalf("创建通义千问客户端失败: %v\n", err)
	}
	fmt.Println("通义千问客户端初始化成功!")

	// 创建上下文
	ctx := context.Background()

	// 示例1: 基本多模态图片分析
	fmt.Println("=== 示例1: 基本多模态图片分析 ===")

	// 使用Unsplash随机图片API获取一张高质量图片
	// 每次请求都会返回不同的随机图片
	fmt.Printf("使用随机图片URL: %s\n\n", NatureImageURL)

	// 创建图像URL内容 - 使用Base64格式
	// 尝试从URL下载图片并转换为Base64
	fmt.Println("正在下载并转换图片为Base64格式...")
	imageURL, err := getImageAsBase64(NatureImageURL, ImageDetailHigh)
	if err != nil {
		log.Fatalf("下载并转换图片失败: %v\n", err)
	}
	fmt.Println("图片转换成功!")

	// 创建多模态消息内容
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "你是一个专业的图像分析助手，擅长分析图像内容并提供详细描述。",
				},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "这张图片是什么？请详细描述一下图片中的内容、场景和氛围。如果有人物，请描述他们的活动。",
				},
				imageURL,
			},
		},
	}

	// 调用多模态API
	fmt.Println("正在分析图片...")
	
	// 打印请求信息
	fmt.Println("请求详情:")
	fmt.Printf("- 模型: %s\n", modelToUse)
	fmt.Println("- 图片URL: [使用Base64编码的图片数据]")
	fmt.Printf("- 图片详细程度: %s\n", imageURL.Detail)
	
	// 添加超时控制 - 增加超时时间
	timeoutCtx, cancel := context.WithTimeout(ctx, 120*time.Second) // 增加到2分钟
	defer cancel()
	
	// 直接调用API
	response, err := llm.GenerateContent(
		timeoutCtx, 
		messages, 
		llms.WithMaxTokens(1000),
		llms.WithTemperature(0.7),
	)
	
	// 错误处理
	if err != nil {
		fmt.Printf("多模态调用出错: %v\n", err)
		fmt.Println("尝试使用不同的参数重试...")
		
		// 使用不同的参数重试
		imageURL.Detail = ImageDetailLow // 降低图片详细程度
		response, err = llm.GenerateContent(
			timeoutCtx, 
			messages, 
			llms.WithMaxTokens(500), // 减少token数量
			llms.WithTemperature(0.5), // 降低温度
		)
		
		if err != nil {
			log.Fatalf("多模态调用失败，重试也失败: %v\n", err)
		}
	}

	// 输出结果
	fmt.Println("多模态回复:")
	if len(response.Choices) > 0 {
		fmt.Println(response.Choices[0].Content)
	}

	// 示例2: 多轮多模态对话
	fmt.Println("\n=== 示例2: 多轮多模态对话 ===")
	// 构建多轮对话消息，包含图片和后续问题
	// 重新使用之前的图片URL内容
	multiTurnMessages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "你是一个专业的图像分析助手，擅长分析图像内容并提供详细描述。",
				},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "这张图片是什么？请简要描述。",
				},
				imageURL, // 使用之前已转换为Base64的图片
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "这是一张自然风景图片。",
				},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "图片中有什么主要颜色？请详细分析一下色彩构成和色调，以及这些颜色如何影响图片的整体氛围。",
				},
			},
		},
	}

	// 调用多轮多模态对话
	fmt.Println("正在进行多轮对话分析...")
	multiTurnResponse, err := llm.GenerateContent(
		ctx,
		multiTurnMessages,
		llms.WithMaxTokens(1000),
		llms.WithTemperature(0.7),
	)
	if err != nil {
		log.Fatalf("多轮多模态对话失败: %v\n", err)
	}

	// 输出结果
	fmt.Println("多轮多模态对话回复:")
	if len(multiTurnResponse.Choices) > 0 {
		fmt.Println(multiTurnResponse.Choices[0].Content)
	}

	// 示例3: 流式多模态输出
	fmt.Println("\n=== 示例3: 流式多模态输出 ===")
	// 使用另一张随机图片 - 城市建筑主题
	fmt.Printf("使用另一个随机图片URL: %s\n\n", ArchitectureImageURL)

	// 创建城市建筑图片URL内容 - 使用Base64格式
	fmt.Println("正在下载并转换城市图片为Base64格式...")
	cityImageURL, err := getImageAsBase64(ArchitectureImageURL, ImageDetailHigh)
	if err != nil {
		log.Fatalf("下载并转换城市图片失败: %v\n", err)
	}
	fmt.Println("城市图片转换成功!")

	// 构建流式输出的多模态消息
	streamMessages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "你是一个专业的建筑分析师，擅长分析建筑风格和城市特点。",
				},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "请分析这张图片并给出详细描述，包括建筑风格、城市特点、建筑材料和整体氛围。如果能识别出具体城市或地标，请一并说明。",
				},
				cityImageURL,
			},
		},
	}

	// 调用流式多模态API
	fmt.Println("正在流式分析图片...")
	fmt.Println("流式多模态输出:")

	// 使用WithStreamingFunc选项实现流式输出
	_, err = llm.GenerateContent(
		ctx,
		streamMessages,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			// 实时打印每个内容片段
			fmt.Print(string(chunk))
			return nil
		}),
		llms.WithMaxTokens(1000),
		llms.WithTemperature(0.7),
	)
	if err != nil {
		log.Fatalf("流式多模态调用失败: %v\n", err)
	}

	fmt.Println("\n")

	// 示例4: 本地图片处理
	fmt.Println("\n=== 示例4: 本地图片处理 ===")
	// 获取本地图片的绝对路径
	localImagePath, err := filepath.Abs(LocalImagePath)
	if err != nil {
		log.Fatalf("获取图片绝对路径失败: %v\n", err)
	}

	// 检查本地图片是否存在
	_, err = os.Stat(localImagePath)
	if os.IsNotExist(err) {
		log.Fatalf("错误: 本地图片文件 %s 不存在\n请确保在示例目录中放置了名为'example.jpg'的图像文件\n", localImagePath)
	}

	fmt.Printf("使用本地图片: %s\n\n", localImagePath)

	// 创建本地图片内容 - 使用base64编码而非file://协议
	// 读取图片文件
	imageData, err := os.ReadFile(localImagePath)
	if err != nil {
		log.Fatalf("读取本地图片失败: %v\n", err)
	}

	// 将图片编码为base64
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	// 创建base64格式的图片内容
	localImage := llms.ImageURLContent{
		URL:    fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64),
		Detail: ImageDetailHigh, // 高详细度
	}

	// 构建本地图片多模态消息
	localImageMessages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "你是一个专业的摄影分析师，擅长分析图片内容、主题和拍摄意图。",
				},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "这张本地图片是什么？请详细描述图片内容、主题和可能的拍摄意图。",
				},
				localImage,
			},
		},
	}

	// 调用本地图片API
	fmt.Println("正在分析本地图片...")
	localImageResponse, err := llm.GenerateContent(
		ctx,
		localImageMessages,
		llms.WithMaxTokens(1000),
		llms.WithTemperature(0.7),
	)
	if err != nil {
		log.Fatalf("本地图片调用失败: %v\n", err)
	}

	// 输出结果
	fmt.Println("本地图片分析结果:")
	if len(localImageResponse.Choices) > 0 {
		fmt.Println(localImageResponse.Choices[0].Content)
	}

	fmt.Println("\n=== 示例运行完成！===")
}
