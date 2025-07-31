package alor

import (
	"encoding/json"
	"sync"
)

type EventType string

const (
	SystemType EventType = "system"
	DataType   EventType = "data"
)

type ChainEvent struct {
	Type   EventType
	Opcode Opcode
	Guid   GUID
	Data   json.RawMessage
	Next   *ChainEvent
}

type ChainQueue struct {
	// Elements  []ChainEvent
	Size      int `json:"size"`
	Len       int `json:"len"`
	mu        sync.Mutex
	firstElem *ChainEvent
	lastElem  *ChainEvent
}

func NewChainQueue(size int) *ChainQueue {
	return &ChainQueue{
		// Elements:  make([]ChainEvent, 0),
		Size:      size,
		Len:       0,
		mu:        sync.Mutex{},
		lastElem:  nil,
		firstElem: nil,
	}
}

func (q *ChainQueue) Enqueue(element *ChainEvent) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.GetLength() == q.Size {
		return ErrQueueOverFlow
	}

	q.Len++

	if q.firstElem == nil {
		q.firstElem = element
		q.lastElem = element
		return nil
	}

	q.lastElem.Next = element
	q.lastElem = element

	return nil
}

func (q *ChainQueue) Dequeue() (*ChainEvent, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.IsEmpty() {
		return nil, ErrQueueUnderFlow
	}

	q.Len--

	element := q.firstElem
	q.firstElem = element.Next

	return element, nil // Slice off the element once it is dequeued.
}

func (q *ChainQueue) GetLength() int {
	// return len(q.Elements)
	return q.Len
}

func (q *ChainQueue) IsEmpty() bool {
	return q.Len <= 0
}
