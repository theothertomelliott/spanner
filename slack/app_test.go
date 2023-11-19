package slack

import (
	"context"
	"testing"
	"time"

	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

func TestHandlerIsCalledForEachEvent(t *testing.T) {
	slackEvents := make(chan socketmode.Event, 10)
	customEvents := make(chan *customEvent, 10)
	runCalls := make(chan struct{}, 1)

	testApp := &app{
		client: runSocketClient{
			runCalls: runCalls,
		},
		slackEvents:   slackEvents,
		combinedEvent: make(chan combinedEvent, 2),
		customEvents:  customEvents,
	}

	results := make(chan struct{}, 2)

	go func() {
		testApp.Run(func(evt spanner.Event) error {
			results <- struct{}{}
			return nil
		})
	}()

	select {
	case <-runCalls:
	case <-time.After(time.Second):
		t.Errorf("expected run to be called on Slack client")
	}

	for i := 0; i < 10; i++ {
		var eventType string
		if i%2 == 0 {
			slackEvents <- socketmode.Event{}
			eventType = "slack"
		} else {
			customEvents <- &customEvent{}
			eventType = "custom"
		}
		select {
		case <-results:
		case <-time.After(time.Second):
			t.Errorf("timeout waiting for %v event", eventType)
		}
	}
}

type runSocketClient struct {
	nilSocketClient

	runCalls chan struct{}
}

func (r runSocketClient) RunContext(context.Context) error {
	r.runCalls <- struct{}{}
	return nil
}

func (runSocketClient) Ack(req socketmode.Request, payload ...interface{}) {}
