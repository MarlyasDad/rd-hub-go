package alor

import (
	"errors"
	"sync"
)

var (
	ErrQueueOverFlow  = errors.New("queue is overflow")
	ErrQueueUnderFlow = errors.New("queue is underflow")
)

type BarQueue struct {
	Elements []*Bar
	Size     int
	Len      int
	mu       sync.Mutex
}

func NewBarQueue(size int) *BarQueue {
	return &BarQueue{
		Elements: make([]*Bar, 0),
		Size:     size,
		Len:      0,
		mu:       sync.Mutex{},
	}
}

func (q *BarQueue) Enqueue(bar *Bar) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.GetLength() == q.Size {
		// return ErrQueueOverFlow
		// Если максимальный размер, удаляем самый старый элемент
		_, _ = q.Dequeue()
	}

	q.Elements = append(q.Elements, bar)
	q.Len++

	return nil
}

func (q *BarQueue) Dequeue() (*Bar, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.IsEmpty() {
		return nil, ErrQueueUnderFlow
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

func (q *BarQueue) GetLength() int {
	// return len(q.Elements)
	return q.Len
}

func (q *BarQueue) IsEmpty() bool {
	return len(q.Elements) == 0
}

func (q *BarQueue) Peek() (*Bar, error) {
	if q.IsEmpty() {
		return nil, errors.New("empty queue")
	}
	return q.Elements[0], nil
}

func (q *BarQueue) GetActiveBar() (*Bar, error) {
	if len(q.Elements) == 0 {
		return nil, errors.New("no available bars")
	}
	return q.Elements[0], nil
}

func (q *BarQueue) GetLastFinalizedBar() (*Bar, error) {
	if len(q.Elements) <= 1 {
		return nil, errors.New("no available bars")
	}
	return q.Elements[1], nil
}

func (q *BarQueue) GetSpecificBar(index int64) *Bar {
	return q.Elements[index]
}

func (q *BarQueue) GetAllBars() []*Bar {
	return q.Elements
}

func (q *BarQueue) GetBarsRange(start, end int64) []*Bar {
	return q.Elements[start:end]
}

func (q *BarQueue) GetHeikenAshiBarsRange(start, end int64) ([]*Bar, error) {
	// Open = (Цена открытия предыдущей свечи + Закрытие предыдущей свечи) / 2
	// Close = (Открытие свечи + Максимальная точка + Минимальная точка + Закрытие свечи) / 4
	// Min = [Минимально значение из (Минимум, Открытие, Закрытие)]
	// Max = [Максимальное значение из (Максимум, Открытие, Закрытие)]

	heikenBars := make([]*Bar, end-start)
	for index, bar := range q.Elements[start:end] {
		if index == 0 {
			heikenBars = append(heikenBars, bar)
			continue
		}
		barOpen := (heikenBars[index-1].Open + heikenBars[index-1].Close) / 2
		barClose := (bar.Open + bar.Close + bar.High + bar.Low) / 4
		barMax := max(bar.Open, bar.Close, bar.High)
		barMin := min(bar.Open, bar.Close, bar.Low)
		newHeikenBar := &Bar{
			Open:   barOpen,
			Close:  barClose,
			High:   barMax,
			Low:    barMin,
			Volume: bar.Volume,
		}
		heikenBars = append(heikenBars, newHeikenBar)
	}

	return heikenBars, nil
}
