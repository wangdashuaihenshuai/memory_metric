package metric

import (
	"time"

	"github.com/wangdashuaihenshuai/memory_metric/state"
)

func newTimer(name string, storage Storage) Timer {
	return &timer{
		name:         name,
		storage:      storage,
		StateMachine: *state.NewStateMachine(),
	}
}

type timer struct {
	state.StateMachine
	storage Storage
	name    string
	tags    Tags
	value   int64
	startAt *time.Time
}

func (t *timer) WithTag(key string, value string) Timer {
	t.tags[key] = value
	return t
}

func (t *timer) Start() {
	now := time.Now()
	t.startAt = &now
}

func (t *timer) End() {
	if t.startAt == nil {
		return
	}

	d := time.Since(*t.startAt)
	t.Value(d.Milliseconds())
}

func (t *timer) Value(value int64) {
	if t.IsState(state.EndState) {
		return
	}

	t.SetState(state.EndState)
	t.value = value
	point := newPoint(t.tags, t.value)
	t.storage.Store(t.name+TimerSumEnd, point)

	point = newPoint(t.tags, 1)
	t.storage.Store(t.name+TimerCountEnd, point)
}
