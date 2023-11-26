package slack

import (
	"testing"

	"github.com/slack-go/slack"
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

	slashCommand := slack.SlashCommand{
		ChannelID: "ABC123",
		UserID:    "DEF456",
		TriggerID: "trigger",
		Command:   "/mycommand",
	}
	slackEvents <- slashCommandEvent(
		slashCommand,
	)

	testApp.Run(func(evt spanner.Event) error {
		defer func() {
			// Stop the client
			close(client.stop)
		}()

		// Expect a slash command with the expected command is received
		cmd := evt.ReceiveSlashCommand(slashCommand.Command)
		if cmd == nil {
			t.Errorf("expected a ReceiveSlashCommand event")
			return nil
		}
		if cmd.Channel().ID() != slashCommand.ChannelID {
			t.Errorf("expected channel id %q, got %q", slashCommand.ChannelID, cmd.Channel().ID())
		}
		if cmd.User().ID() != slashCommand.UserID {
			t.Errorf("expected user id %q, got %q", slashCommand.UserID, cmd.User().ID())
		}

		return nil
	})

}
