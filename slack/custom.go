package slack

import "github.com/theothertomelliott/spanner"

var _ spanner.CustomEvent = &customEvent{}

func NewCustomEvent(body map[string]interface{}) spanner.CustomEvent {
	return &customEvent{
		MessageSender: &MessageSender{},
		body:          body,
	}
}

type customEvent struct {
	*MessageSender `json:"ms"`

	body map[string]interface{}
}

func (c *customEvent) Body() map[string]interface{} {
	return c.body
}
