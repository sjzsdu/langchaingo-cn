package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cnllms "github.com/sjzsdu/langchaingo-cn/llms"
)

func main() {
	llm := ""
	if len(os.Args) > 1 {
		llm = os.Args[1]
	}

	// 初始化所有可用于 embedding 的模型
	models, modelNames, err := cnllms.InitEmbeddingModels(llm)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	texts := []string{
		"Hello world",
		"LangChain Go is great",
		"向量检索可以加速语义搜索",
	}

	for i, m := range models {
		fmt.Printf("\n===== 使用 %s Embedding =====\n", modelNames[i])
		embs, err := m.EmbedDocuments(ctx, texts)
		if err != nil {
			fmt.Printf("使用 %s 生成向量失败: %v\n", modelNames[i], err)
			continue
		}
		if len(embs) == 0 {
			fmt.Printf("%s 返回空向量\n", modelNames[i])
			continue
		}
		fmt.Printf("输入条数: %d, 向量维度: %d\n", len(embs), len(embs[0]))
	}
}
