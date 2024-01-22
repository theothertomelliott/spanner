package slack

import (
	"fmt"

	"github.com/slack-go/slack"
	"github.com/theothertomelliott/spanner"
)

func renderSlackError(err error) error {
	if err == nil {
		return nil
	}

	if slackErr, ok := err.(slack.SlackErrorResponse); ok {
		return fmt.Errorf("%w %v %v", slackErr, slackErr.ResponseMetadata.Messages, slackErr.ResponseMetadata.Warnings)
	}
	return err
}

var _ spanner.ErrorEvent = &errorEvent{}

func newErrorEvent(err error) *errorEvent {
	q := &actionQueue{}
	return &errorEvent{
		actionQueue: q,
		sender: &MessageSender{
			actionQueue: q,
		},
		err: err,
	}
}

type errorEvent struct {
	actionQueue *actionQueue
	sender      *MessageSender

	err error
}

func (e *errorEvent) SendMessage(channelID string) spanner.ErrorMessage {
	return e.sender.SendMessage(channelID)
}

// ReceiveError implements spanner.ErrorEvent.
func (e *errorEvent) ReceiveError() error {
	return e.err
}
