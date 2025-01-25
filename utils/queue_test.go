package utils

import (
	"errors"
	"testing"
)

func TestQueueImpl_Enqueue(t *testing.T) {
	queue := NewQueue[string]()

	queue.Enqueue("a")
	queue.Enqueue("b")

	if queue.IsEmpty() {
		t.Fatalf("queue is empty")
	}

	if queue.Size() != 2 {
		t.Fatalf("queue has size %d, want %d", queue.Size(), 2)
	}
}

func TestQueueImpl_Dequeue(t *testing.T) {
	queue := NewQueue[string]()

	queue.Enqueue("a")
	queue.Enqueue("b")

	if queue.IsEmpty() {
		t.Fatalf("queue is empty")
	}

	if queue.Size() != 2 {
		t.Fatalf("queue has size %d, want %d", queue.Size(), 2)
	}

	item, err := queue.Dequeue()
	if err != nil {
		t.Fatalf("queue dequeue failed: %v", err)
	}

	if item == nil {
		t.Fatalf("queue dequeue returned nil")
	}

	if *item != "a" {
		t.Fatalf("queue dequeue got %v, want %v", item, "a")
	}

	item, err = queue.Dequeue()
	if err != nil {
		t.Fatalf("queue dequeue failed: %v", err)
	}

	if item == nil {
		t.Fatalf("queue dequeue returned nil")
	}

	if *item != "b" {
		t.Fatalf("queue dequeue got %v, want %v", item, "b")
	}

	_, err = queue.Dequeue()

	if !errors.Is(err, ErrQueueEmpty) {
		t.Fatalf("queue dequeue got %v, want %v", err, ErrQueueEmpty)
	}
}

func TestQueueImpl_RemoveFunc(t *testing.T) {
	queue := NewQueue[string]()
	queue.Enqueue("a")
	queue.Enqueue("b")

	queue.RemoveFunc(func(s string) bool {
		return s == "b"
	})

	if queue.IsEmpty() {
		t.Fatalf("queue is empty")
	}

	if queue.Size() != 1 {
		t.Fatalf("queue has size %d, want %d", queue.Size(), 1)
	}
}

func TestQueueImpl_Items(t *testing.T) {
	queue := NewQueue[string]()
	queue.Enqueue("a")
	queue.Enqueue("b")

	items := queue.Items()
	if len(items) != 2 {
		t.Fatalf("queue has size %d, want %d", len(items), 2)
	}

	if items[0] != "a" {
		t.Fatalf("queue got %v, want %v", items[0], "a")
	}

	if items[1] != "b" {
		t.Fatalf("queue got %v, want %v", items[1], "b")
	}
}
