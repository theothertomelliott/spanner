package slack

import (
	"testing"
	"time"

	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

func TestHandlerIsCalledForEachEvent(t *testing.T) {
	client := newTestClient()
	testApp := client.CreateApp()

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
			client.SendEventToApp(socketmode.Event{})
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
