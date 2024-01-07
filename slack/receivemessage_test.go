package slack

import (
	"context"
	"testing"

	"github.com/slack-go/slack/slackevents"
	"github.com/theothertomelliott/spanner"
)

func TestReceiveMessageContent(t *testing.T) {
	client := newTestClient()
	testApp := client.CreateApp()

	message := slackevents.MessageEvent{
		Channel: "ABC123",
		User:    "DEF456",
		Text:    "hello, there",
	}
	client.SendEventToAppAsync(messageEvent(message))

	testApp.Run(func(ctx context.Context, evt spanner.Event) error {
		defer func() {
			// Stop the client
			close(client.stop)
		}()

		// Expect a message with the right channel id attached
		msg := evt.ReceiveMessage()
		if msg == nil {
			t.Errorf("expected a ReceiveMessage event")
			return nil
		}
		if msg.Channel().ID() != message.Channel {
			t.Errorf("expected channel id %q, got %q", message.Channel, msg.Channel().ID())
		}
		if msg.User().ID() != message.User {
			t.Errorf("expected user id %q, got %q", message.User, msg.User().ID())
		}
		if msg.Text() != message.Text {
			t.Errorf("expected text %q, got %q", message.Text, msg.Text())
		}

		return nil
	})

}
