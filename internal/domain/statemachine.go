package domain

import "context"

type StateMachine[P any] struct {
	currentState StateFunc[P]
}

func NewStateMachine[P any](initial StateFunc[P]) StateMachine[P] {
	return StateMachine[P]{
		currentState: initial,
	}
}

func (sm *StateMachine[P]) Execute(ctx context.Context, payload P) error {
	next, err := sm.currentState(ctx, payload)
	if err != nil {
		return err
	}
	if next != nil {
		sm.currentState = next
	}
	return nil
}

func (sm *StateMachine[P]) Set(state StateFunc[P]) {
	sm.currentState = state
}

// StateFunc is a function type that accepts context and request, returns next state function and error.
type StateFunc[P any] func(ctx context.Context, payload P) (StateFunc[P], error)
