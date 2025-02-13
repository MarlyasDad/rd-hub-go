package alor

import (
	"errors"
	"sync"
)

var (
	ErrQueueOverFlow  = errors.New("queue is overflow")
	ErrQueueUnderFlow = errors.New("queue is underflow")
)

type Queue struct {
	Elements []Event
	Size     int
	Len      int
	mu       sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		Elements: make([]Event, 0),
		Size:     0,
		Len:      0,
		mu:       sync.Mutex{},
	}
}

func (q *Queue) Enqueue(elem Event) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.GetLength() == q.Size {
		return ErrQueueOverFlow
	}
	q.Elements = append(q.Elements, elem)
	q.Len++

	return nil
}

func (q *Queue) Dequeue() (Event, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.IsEmpty() {
		return Event{}, ErrQueueUnderFlow
	}
	element := q.Elements[0]
	if q.GetLength() == 1 {
		q.Elements = nil
		return element, nil
	}
	q.Elements = q.Elements[1:]
	q.Len--
	return element, nil // Slice off the element once it is dequeued.
}

func (q *Queue) GetLength() int {
	// return len(q.Elements)
	return q.Len
}

func (q *Queue) IsEmpty() bool {
	return len(q.Elements) == 0
}

func (q *Queue) Peek() (Event, error) {
	if q.IsEmpty() {
		return Event{}, errors.New("empty queue")
	}
	return q.Elements[0], nil
}

// Обработчик очереди сюда!
