package utils

import (
	"sync"

	"github.com/avakumov/metrics/internal/logger"
	"go.uber.org/zap"
)

// WorkerPool представляет пул воркеров
type WorkerPool struct {
	tasks   chan int // Канал для задач
	wg      sync.WaitGroup
	workers int
	fn      func()
}

// NewWorkerPool создает новый пул воркеров
func NewWorkerPool(workers int, bufferSize int, fn func()) *WorkerPool {
	return &WorkerPool{
		tasks:   make(chan int, bufferSize),
		workers: workers,
		fn:      fn,
	}
}

// Start запускает воркеров
func (p *WorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker выполняет задачи
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	for task := range p.tasks {
		logger.Log.Debug("Send metrics by worker in task:", zap.Int("worker", id), zap.Int("task", task))
		p.fn()
	}
}

// Stop завершает работу пула
func (p *WorkerPool) Stop() {
	close(p.tasks) // Закрываем канал задач
	p.wg.Wait()    // Ждем завершения всех воркеров
}

// Submit добавляет задачу в очередь
func (p *WorkerPool) Submit(task int) {
	p.tasks <- task
}
