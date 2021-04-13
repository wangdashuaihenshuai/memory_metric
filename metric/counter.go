package metric

import (
	"github.com/wangdashuaihenshuai/memory_metric/state"
)

func newCounter(name string, storage Storage) Counter {
	return &counter{
		name:         name,
		storage:      storage,
		StateMachine: *state.NewStateMachine(),
	}
}

type counter struct {
	state.StateMachine
	storage Storage
	name    string
	tags    Tags
	value   int64
}

func (c *counter) WithTag(key string, value string) Counter {
	c.tags[key] = value
	return c
}

func (t *counter) Value(value int64) {
	if t.IsState(state.EndState) {
		return
	}

	t.SetState(state.EndState)
	t.value = value
	point := newPoint(t.tags, t.value)
	t.storage.Store(t.name, point)
}

func (t *counter) Count() {
	t.Value(1)
}
