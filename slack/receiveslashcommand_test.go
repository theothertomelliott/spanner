package slack

import (
	"context"
	"testing"

	"github.com/slack-go/slack"
	"github.com/theothertomelliott/spanner"
)

func TestReceiveSlashCommand(t *testing.T) {
	client := newTestClient([]string{"ABC123"})
	testApp := client.CreateApp()

	slashCommand := slack.SlashCommand{
		ChannelID: "ABC123",
		UserID:    "DEF456",
		TriggerID: "trigger",
		Command:   "/mycommand",
	}
	client.SendEventToAppAsync(slashCommandEvent(
		slashCommand,
	))

	testApp.Run(func(ctx context.Context, evt spanner.Event) error {
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
