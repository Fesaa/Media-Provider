package utils

import "errors"

type Queue[T any] interface {
	Enqueue(item T)
	Dequeue() (*T, error)
	IsEmpty() bool
}

type queueImpl[T any] struct {
	items []T
}

func NewQueue[T any]() Queue[T] {
	return &queueImpl[T]{items: []T{}}
}

func (q *queueImpl[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

func (q *queueImpl[T]) Dequeue() (*T, error) {
	if q.IsEmpty() {
		return nil, errors.New("queue is empty")
	}
	item := q.items[0]
	q.items = q.items[1:]
	return &item, nil
}

func (q *queueImpl[T]) IsEmpty() bool {
	return len(q.items) == 0
}
