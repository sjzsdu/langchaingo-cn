# Makefile for langchaingo-cn
# 此文件包含项目的便捷目标，不作为构建系统使用
# 详见README获取更多信息

.PHONY: help test fmt lint clean deps update-deps test-race test-cover build-examples

# 默认目标
all: fmt test

# 显示帮助信息
help:
	@echo "可用目标："
	@echo ""
	@echo "测试："
	@echo "  test           - 运行所有测试"
	@echo "  test-race      - 运行带竞态检测的测试"
	@echo "  test-cover     - 运行带覆盖率报告的测试"
	@echo ""
	@echo "代码质量："
	@echo "  fmt            - 格式化代码"
	@echo "  lint           - 运行代码检查"
	@echo "  lint-fix       - 运行代码检查并自动修复问题"
	@echo ""
	@echo "示例："
	@echo "  run-examples   - 运行所有示例"
	@echo "  run-qwen-example - 运行通义千问示例"
	@echo "  run-qwen-multimodal-example - 运行通义千问多模态示例"
	@echo "  run-qwen-openai-example - 运行通义千问OpenAI兼容示例"
	@echo "  run-deepseek-example - 运行DeepSeek示例"
	@echo "  run-deepseek-multimodal-example - 运行DeepSeek多模态示例"
	@echo "  run-deepseek-streaming-example - 运行DeepSeek流式输出示例"
	@echo "  run-deepseek-tool-call-example - 运行DeepSeek工具调用示例"
	@echo "  build-examples - 构建所有示例项目"
	@echo ""
	@echo "其他："
	@echo "  deps           - 下载依赖"
	@echo "  update-deps    - 更新依赖"
	@echo "  clean          - 清理生成的文件"
	@echo "  help           - 显示此帮助信息"

# 运行所有测试
test:
	go test -v ./...

# 格式化代码
fmt:
	go fmt ./...

# 运行代码检查
lint: lint-deps
	golangci-lint run --color=always ./...

# 运行代码检查并自动修复问题
lint-fix:
	golangci-lint run --fix ./...

# 安装lint依赖
lint-deps:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo >&2 "golangci-lint未找到，正在安装..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}

# 清理生成的文件
clean: clean-lint-cache
	go clean

# 清理lint缓存
clean-lint-cache:
	golangci-lint cache clean

# 下载依赖
deps:
	go mod download

# 更新依赖
update-deps:
	go get -u ./...
	@echo "更新go.mod文件..."
	go mod tidy

# 运行带竞态检测的测试
test-race:
	go test -race ./...

# 运行带覆盖率报告的测试
test-cover:
	go test -cover ./...

# 构建所有示例项目
build-examples:
	@echo "构建所有示例项目..."
	for example in $$(find ./examples -mindepth 1 -maxdepth 1 -type d); do \
		(cd $$example && echo "构建 $$example" && go mod tidy && go build -o /dev/null) || exit 1; \
	done