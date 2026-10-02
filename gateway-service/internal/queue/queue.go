package queue

import (
	"context"
	"log"
	"time"
)

type TaskHandler func(ctx context.Context, task Task) error

type Queue struct {
	tasks   chan Task
	handler TaskHandler
}

func New(bufferSize int, handler TaskHandler) *Queue{
	return &Queue{
		tasks:   make(chan Task, bufferSize),
		handler: handler,
	}
}

func (q *Queue) Enqueue(task Task) {
	if task.MaxRetries == 0 {
		task.MaxRetries = 5
	}
	select {
	case q.tasks <- task:
		log.Printf("task enqueued: type=%s payload=%v", task.Type, task.Payload)
	default:
		log.Printf("queue full, task dropped: type=%s", task.Type)
	}
}

func (q *Queue) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("queue worker stopped")
				return
			case task := <-q.tasks:
				log.Printf("processing task: type=%s retry=%d", task.Type, task.Retries)
				if err := q.handler(ctx, task); err != nil {
					task.Retries++
					if task.Retries < task.MaxRetries {
						log.Printf("task failed, will retry in 5s: %v", err)
						time.AfterFunc(5*time.Second, func() {
							q.Enqueue(task)
						})
					} else {
						log.Printf("task failed after %d retries: type=%s", task.Retries, task.Type)
					}
				} else {
					log.Printf("task completed: type=%s", task.Type)
				}
			}
		}
	}()
}