.PHONY: help test format 

help:
	@echo "  test           Запустить тесты"
	@echo "  format         Форматировать весь проект"
	@echo "  build          build agent, server"
	@echo "  build-client   build agent"
	@echo "  build-server   build server"
	@echo "  clean          clean binary build"
	@echo "  lint           lint by golangci-lint"
	@echo "  dev-server     start with debug logger"
	@echo "  dev-client     start with debug logger"


test:
	gotest -v ./...

format:
	go fmt	./...


all: build

build: build-client build-server

build-client:
	@echo "Building client..."
	go build -o cmd/agent/main ./cmd/agent

build-server:
	@echo "Building server..."
	go build -o cmd/server/main ./cmd/server

clean:
	@echo "Cleaning..."
	rm -f cmd/server/main cmd/agent/main

lint:
	@echo "Linting..."
	golangci-lint run

dev-server:
	@echo "Run dev server ..."
	go run ./cmd/server/main.go -l=debug

dev-client:
	@echo "Run dev client ..."
	go run ./cmd/agent/main.go -l=debug
