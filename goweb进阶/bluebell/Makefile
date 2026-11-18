.DEFAULT_GOAL := help

APP_NAME := bluebell
BUILD_DIR := bin

.PHONY: all build run test fmt vet check gotool clean help

all: check build

## 编译项目，输出到 bin/bluebell.exe
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME).exe .

## 直接运行项目（需要先启动 MySQL 和 Redis）
run:
	go run .

## 运行全部测试
test:
	go test ./...

## 格式化 Go 源码
fmt:
	gofmt -w $$(go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...)

## 运行 go vet 静态检查
vet:
	go vet ./...

## 执行格式化和静态检查
gotool: fmt vet

## 执行测试和静态检查
check: gotool test

## 删除构建产物和 Go 缓存
clean:
	go clean -cache
	rm -rf $(BUILD_DIR)

## 显示可用命令
help:
	@printf "可用命令：\n"
	@printf "  make build  - 编译项目到 $(BUILD_DIR)/$(APP_NAME).exe\n"
	@printf "  make run    - 运行项目\n"
	@printf "  make test   - 运行全部测试\n"
	@printf "  make fmt    - 格式化 Go 源码\n"
	@printf "  make vet    - 执行静态检查\n"
	@printf "  make check  - 执行格式化、静态检查和测试\n"
	@printf "  make clean  - 删除构建产物和缓存\n"
