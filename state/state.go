package state

import "sync/atomic"

type State int64

const (
	InitState  = 0
	StartState = 1
	EndState   = 2
)

type StateMachine struct {
	state State
}

func (s *StateMachine) IsState(state State) bool {
	loadState := State(atomic.LoadInt64((*int64)(&s.state)))
	return loadState == state
}

func (s *StateMachine) SetState(state State) {
	atomic.StoreInt64((*int64)(&s.state), int64(state))
}

func (s *StateMachine) GotoState(state State, oldState State) bool {
	return atomic.CompareAndSwapInt64((*int64)(&s.state), int64(oldState), int64(state))
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		state: InitState,
	}
}
