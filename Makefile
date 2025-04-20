APP=gdss

.PHONY: build
build:
	@go build -o ${APP} main.go

.PHONY: run
run:
	go run -race main.go

test:
	@echo "Running tests..."
	go test ./... -v
	# go test -v main.go

.PHONY: clean
clean:
	go clean
	
gotool:
	go fmt ./
	go vet ./

help:
	@echo "make - 格式化 Go 代码, 并编译生成二进制文件"
	@echo "make build - 编译 Go 代码, 生成二进制文件"
	@echo "make run - 直接运行 Go 代码"
	@echo "make clean - 移除二进制文件和 vim swap files"
	@echo "make gotool - 运行 Go 工具 'fmt' and 'vet'"
