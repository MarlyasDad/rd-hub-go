package alor

import (
	"sync"
)

type Counter struct {
	mut        sync.Mutex
	processing int64
	total      int64
}

func (t *Counter) Add() {
	t.mut.Lock()
	defer t.mut.Unlock()

	t.processing++
	t.total++
}

func (t *Counter) Done() {
	t.mut.Lock()
	defer t.mut.Unlock()

	t.processing--
}

func (t *Counter) Processing() int64 {
	return t.processing
}

func (t *Counter) Accumulated() int64 {
	t.mut.Lock()
	defer t.mut.Unlock()

	var total int64
	total = t.total
	t.total = 0

	return total
}
