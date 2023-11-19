package slack

import (
	"testing"

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

	expectedChannelID := "ABC123"
	expectedUserID := "DEF456"
	expectedText := "hello, there"
	slackEvents <- messageEvent(
		expectedChannelID,
		expectedUserID,
		expectedText,
	)

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
		if msg.Channel().ID() != expectedChannelID {
			t.Errorf("expected channel id %q, got %q", expectedChannelID, msg.Channel().ID())
		}
		if msg.User().ID() != expectedUserID {
			t.Errorf("expected user id %q, got %q", expectedChannelID, msg.User().ID())
		}
		if msg.Text() != expectedText {
			t.Errorf("expected text %q, got %q", expectedText, msg.Text())
		}

		return nil
	})

}
