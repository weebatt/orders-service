package faulttolerance

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func Retry(operation func() error, maxRetries int, baseDelay int64) error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}
		sleepTime := time.Duration(baseDelay<<attempt) * time.Millisecond
		time.Sleep(sleepTime)
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

func Timeout(operation func() error, timeoutMs int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		errCh <- operation()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return errors.New("operation timed out")
	}
}

type DeadLetterQueue struct {
	mu       sync.Mutex
	messages []string
}

func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		messages: make([]string, 0),
	}
}

func (dlq *DeadLetterQueue) Add(msg string) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.messages = append(dlq.messages, msg)
}

func (dlq *DeadLetterQueue) GetMessages() []string {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	result := make([]string, len(dlq.messages))
	copy(result, dlq.messages)
	return result
}

func ProcessWithDLQ(messages []string, handler func(string) error, dlq *DeadLetterQueue) {
	for _, msg := range messages {
		err := handler(msg)
		if err != nil {
			dlq.Add(msg)
		}
	}
}
