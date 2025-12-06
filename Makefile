.PHONY: help test format 

help:
	@echo "  test           Запустить тесты"
	@echo "  format         Форматировать весь проект"
	@echo "  build          build agent, server"
	@echo "  build-client   build agent"
	@echo "  build-server   build server"
	@echo "  clean          clean binary build"


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
