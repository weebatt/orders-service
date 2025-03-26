package faulttolerance

import (
	"errors"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	attempts := 0
	op := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("failed attempt")
		}
		return nil
	}

	err := Retry(op, 5, 100)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_Failure(t *testing.T) {
	op := func() error {
		return errors.New("always fail")
	}
	err := Retry(op, 3, 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTimeout_Success(t *testing.T) {
	op := func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}
	err := Timeout(op, 500)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestTimeout_Timeout(t *testing.T) {
	op := func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}
	err := Timeout(op, 100)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestProcessWithDLQ(t *testing.T) {
	messages := []string{"msg1", "msg2", "msg3", "msg4"}
	dlq := NewDeadLetterQueue()

	handler := func(msg string) error {
		if msg == "msg2" || msg == "msg4" {
			return errors.New("fail on msg2/msg4")
		}
		return nil
	}

	ProcessWithDLQ(messages, handler, dlq)
	dlqMessages := dlq.GetMessages()
	if len(dlqMessages) != 2 {
		t.Fatalf("expected 2 messages in DLQ, got %d", len(dlqMessages))
	}
	if dlqMessages[0] != "msg2" || dlqMessages[1] != "msg4" {
		t.Errorf("expected [msg2, msg4], got %v", dlqMessages)
	}
}
