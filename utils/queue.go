package utils

import (
	"errors"
	"slices"
)

type Queue[T comparable] interface {
	Enqueue(item T)
	Dequeue() (*T, error)
	IsEmpty() bool
	Size() int
	Items() []T
	RemoveFunc(func(T) bool) bool
}

type queueImpl[T comparable] struct {
	items []T
}

func NewQueue[T comparable]() Queue[T] {
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

func (q *queueImpl[T]) Size() int {
	return len(q.items)
}

func (q *queueImpl[T]) Items() []T {
	return slices.Clone(q.items)
}

func (q *queueImpl[T]) RemoveFunc(f func(T) bool) bool {
	size := q.Size()
	q.items = slices.DeleteFunc(q.items, f)
	return size != q.Size()
}
