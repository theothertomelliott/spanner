package slack

import (
	"context"

	"github.com/theothertomelliott/spanner"
)

var _ spanner.EphemeralSender = &ephemeralSender{}
var _ eventPopulator = &ephemeralSender{}

type ephemeralSender struct {
	actionQueue *actionQueue

	Text *string `json:"ephemeral"`
}

// SendEphemeralMessage implements spanner.EphemeralSender.
func (es *ephemeralSender) SendEphemeralMessage(text string) {
	es.actionQueue.enqueue(&sendEphemeralMessageAction{
		text: text,
	})
}

func (es *ephemeralSender) populateEvent(ctx context.Context, p eventPopulation, depth int) error {
	return nil
}

var _ action = &sendEphemeralMessageAction{}

type sendEphemeralMessageAction struct {
	text string

	errFunc spanner.ErrorFunc
}

func (e *sendEphemeralMessageAction) ErrorFunc(ef spanner.ErrorFunc) {
	e.errFunc = ef
}

func (e *sendEphemeralMessageAction) getErrorFunc() spanner.ErrorFunc {
	return e.errFunc
}

// Data implements action.
func (e *sendEphemeralMessageAction) Data() interface{} {
	// TODO: This should be more well-defined
	return map[string]interface{}{
		"text": e.text,
	}
}

// Type implements action.
func (*sendEphemeralMessageAction) Type() string {
	return "ephemeral-message"
}

// exec implements action.
func (e *sendEphemeralMessageAction) exec(ctx context.Context, req request) (interface{}, error) {
	payload := map[string]interface{}{
		"text": e.text,
	}
	return payload, nil
}
