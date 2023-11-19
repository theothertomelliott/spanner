package slack

import (
	"testing"

	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

func TestReceiveSlashCommand(t *testing.T) {
	slackEvents := make(chan socketmode.Event, 10)
	client := newRunSocketClient()

	testApp := newAppWithClient(
		client,
		slackEvents,
	)

	expectedChannelID := "ABC123"
	expectedUserID := "DEF456"
	expectedCommand := "/mycommand"
	slackEvents <- slashCommandEvent(
		expectedChannelID,
		expectedUserID,
		"trigger",
		expectedCommand,
	)

	testApp.Run(func(evt spanner.Event) error {
		defer func() {
			// Stop the client
			close(client.stop)
		}()

		// Expect a slash command with the expected command is received
		cmd := evt.ReceiveSlashCommand(expectedCommand)
		if cmd == nil {
			t.Errorf("expected a ReceiveSlashCommand event")
			return nil
		}
		if cmd.Channel().ID() != expectedChannelID {
			t.Errorf("expected channel id %q, got %q", expectedChannelID, cmd.Channel().ID())
		}
		if cmd.User().ID() != expectedUserID {
			t.Errorf("expected user id %q, got %q", expectedChannelID, cmd.User().ID())
		}

		return nil
	})

}
