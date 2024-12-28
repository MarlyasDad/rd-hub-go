package ops_counter

import (
	"context"
	"sync"
	"time"
)

func NewCounter(ctx context.Context) Counter {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	return Counter{
		ctx:    ctx,
		cancel: cancel,
	}
}

type Counter struct {
	ctx     context.Context
	cancel  context.CancelFunc
	Targets map[string]*Target
}

func (c Counter) NewTarget(name string, avg int64, limit int64) *Target {
	target := NewTarget(name, avg, limit)
	c.Targets[name] = &target

	return &target
}

func (c Counter) Start(ctx context.Context) {
	for _, target := range c.Targets {
		go target.Start(ctx)
	}
}

func (c Counter) Stop() {
	c.cancel()
}

func (c Counter) ShowAll() map[string]int64 {
	summary := map[string]int64{}

	for name, target := range c.Targets {
		summary[name] = target.Show()
	}

	return summary
}

func NewTarget(name string, avg int64, limit int64) Target {
	return Target{
		name:     name,
		avg:      avg,
		limit:    limit,
		position: 0,
		stats:    make([]int64, avg, 0),
		counter:  0,
		total:    0,
	}
}

type Target struct {
	name     string
	avg      int64
	limit    int64
	position int64
	stats    []int64
	counter  int64
	mut      sync.Mutex
	total    int64
}

func (t *Target) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Second):
		t.Fix()
	}
}

func (t *Target) Fix() {
	t.mut.Lock()
	defer t.mut.Unlock()

	t.total -= t.stats[t.position]
	t.total += t.stats[t.counter]

	t.stats[t.position] = t.counter
	t.ChangePosition()
	t.counter = 0
}

func (t *Target) Increment() {
	t.mut.Lock()
	defer t.mut.Unlock()

	t.counter++
}

func (t *Target) ChangePosition() {
	if t.position < t.avg-1 {
		t.position++
	} else {
		t.position = 0
	}
}

func (t *Target) Show() int64 {
	return t.total / t.avg
}
