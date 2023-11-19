package slack

import (
	"context"
	"testing"
	"time"

	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

func TestHandlerIsCalledForEachEvent(t *testing.T) {
	slackEvents := make(chan socketmode.Event)
	combinedEvents := make(chan combinedEvent, 2)
	testApp := &app{
		client:        runSocketClient{},
		slackEvents:   slackEvents,
		combinedEvent: combinedEvents,
		customEvents:  make(chan *customEvent, 2),
	}

	results := make(chan struct{}, 2)

	go func() {
		testApp.Run(func(evt spanner.Event) error {
			results <- struct{}{}
			return nil
		})
	}()

	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			slackEvents <- socketmode.Event{}
		} else {
			combinedEvents <- combinedEvent{}
		}
		select {
		case <-results:
		case <-time.After(time.Second):
			t.Errorf("timeout waiting for event")
		}
	}
}

type runSocketClient struct {
	nilSocketClient
}

func (runSocketClient) RunContext(context.Context) error {
	return nil
}

func (runSocketClient) Ack(req socketmode.Request, payload ...interface{}) {}
