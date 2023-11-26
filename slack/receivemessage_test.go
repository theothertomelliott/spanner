package slack

import (
	"testing"

	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

func TestReceiveMessageContent(t *testing.T) {
	slackEvents := make(chan socketmode.Event, 10)
	client := newRunSocketClient()

	testApp := newAppWithClient(
		client,
		slackEvents,
	)

	message := slackevents.MessageEvent{
		Channel: "ABC123",
		User:    "DEF456",
		Text:    "hello, there",
	}
	slackEvents <- messageEvent(message)

	testApp.Run(func(evt spanner.Event) error {
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
