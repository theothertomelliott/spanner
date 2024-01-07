package slack

import (
	"context"

	"github.com/theothertomelliott/spanner"
)

var _ spanner.EphemeralSender = &ephemeralSender{}
var _ eventPopulator = &ephemeralSender{}
var _ eventFinisher = &ephemeralSender{}

type ephemeralSender struct {
	Text *string `json:"ephemeral"`
}

// SendEphemeralMessage implements spanner.EphemeralSender.
func (es *ephemeralSender) SendEphemeralMessage(text string) {
	es.Text = &text
}

func (es *ephemeralSender) finishEvent(ctx context.Context, req request) error {
	payload := map[string]interface{}{
		"text": es.Text,
	}
	req.client.Ack(req.req, payload)

	return nil
}

func (es *ephemeralSender) populateEvent(ctx context.Context, p eventPopulation, depth int) error {
	return nil
}
