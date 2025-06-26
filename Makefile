# Makefile for langchaingo-cn

.PHONY: test fmt lint clean run-example

# 默认目标
all: fmt test

# 运行所有测试
test:
	go test -v ./...

# 格式化代码
fmt:
	go fmt ./...

# 运行代码检查
lint:
	golangci-lint run

# 清理生成的文件
clean:
	go clean

# 运行DeepSeek示例
run-deepseek-example:
	go run examples/deepseek/main.go

# 下载依赖
deps:
	go mod download

# 更新依赖
update-deps:
	go get -u ./...
	@echo "更新go.mod文件..."
	go mod tidy