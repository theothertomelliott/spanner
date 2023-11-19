package slack

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

func TestHandlerIsCalledForEachEvent(t *testing.T) {
	slackEvents := make(chan socketmode.Event, 10)
	client := newRunSocketClient()

	testApp := newAppWithClient(
		client,
		slackEvents,
	)

	results := make(chan struct{}, 2)

	go func() {
		testApp.Run(func(evt spanner.Event) error {
			results <- struct{}{}
			return nil
		})
	}()

	for i := 0; i < 10; i++ {
		var eventType string
		if i%2 == 0 {
			slackEvents <- socketmode.Event{}
			eventType = "slack"
		} else {
			testApp.SendCustom(NewCustomEvent(map[string]interface{}{}))
			eventType = "custom"
		}
		select {
		case <-results:
		case <-time.After(time.Second):
			t.Errorf("timeout waiting for %v event", eventType)
		}
	}

	if client.runCount != 1 {
		t.Errorf("expected run to be called exactly once")
	}
}

func newRunSocketClient() *runSocketClient {
	return &runSocketClient{
		stop: make(chan struct{}),
	}
}

type runSocketClient struct {
	nilSocketClient

	runMtx   sync.Mutex
	runCount int

	stop chan struct{}
}

func (r *runSocketClient) RunContext(context.Context) error {
	r.runMtx.Lock()
	r.runCount++
	r.runMtx.Unlock()
	<-r.stop // Must block
	return nil
}

func (*runSocketClient) Ack(req socketmode.Request, payload ...interface{}) {}

func messageEvent(channelID, userID, text string) socketmode.Event {
	return socketmode.Event{
		Type: socketmode.EventTypeEventsAPI,
		Data: slackevents.EventsAPIEvent{
			Type: slackevents.CallbackEvent,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &slackevents.MessageEvent{
					Text:    text,
					User:    userID,
					Channel: channelID,
				},
			},
		},
	}
}

func slashCommandEvent(channelID, userID, triggerID, command string) socketmode.Event {
	return socketmode.Event{
		Type: socketmode.EventTypeSlashCommand,
		Data: slack.SlashCommand{
			ChannelID: channelID,
			UserID:    userID,
			TriggerID: triggerID,
			Command:   command,
		},
	}
}
