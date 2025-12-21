package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func runCommand(ctx context.Context, dir, name string, args ...string) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = dir

	fmt.Printf("Starting: %s %v\n", name, args)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Command failed: %v\n", err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем сервер в горутине
	go func() {
		runCommand(ctx, "cmd/server", "go", "run", ".", "-l", "debug")
	}()

	// Даем серверу время запуститься
	time.Sleep(2 * time.Second)

	// Запускаем клиент
	go func() {
		runCommand(ctx, "cmd/agent", "go", "run", ".", "-l", "debug")
	}()

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Shutting down...")
	cancel()
	time.Sleep(1 * time.Second)
}
