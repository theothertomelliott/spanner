package slack

import (
	"context"

	"github.com/theothertomelliott/spanner"
)

var _ spanner.CustomEvent = &customEvent{}

func NewCustomEvent(body map[string]interface{}) spanner.CustomEvent {
	return &customEvent{
		body: body,
	}
}

type customEvent struct {
	ctx  context.Context
	body map[string]interface{}
}

func (c *customEvent) Body() map[string]interface{} {
	return c.body
}
