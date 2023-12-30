package slack

import (
	"github.com/theothertomelliott/spanner"
)

var _ spanner.EphemeralSender = &ephemeralSender{}

type ephemeralSender struct {
	Text *string `json:"ephemeral"`
}

// SendEphemeralMessage implements spanner.EphemeralSender.
func (es *ephemeralSender) SendEphemeralMessage(text string) {
	es.Text = &text
}

func (es *ephemeralSender) finishEvent(req request) error {
	payload := map[string]interface{}{
		"text": es.Text,
	}
	req.client.Ack(req.req, payload)

	return nil
}

func (es *ephemeralSender) populateEvent(p eventPopulation) error {
	return nil
}
