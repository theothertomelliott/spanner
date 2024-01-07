package slack

import (
	"context"

	"github.com/theothertomelliott/spanner"
)

type action interface {
	spanner.Action

	// exec performs and action and returns a payload to acknowledge the request as appropriate
	exec(ctx context.Context, req request) (interface{}, error)
}

type actionQueue struct {
	actions []action
}

func (a *actionQueue) enqueue(ac action) {
	a.actions = append(a.actions, ac)
}
