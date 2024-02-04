package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/theothertomelliott/spanner"
)

// TestGettingStarted verifies that the code in examples/gettingstarted
// interacts with Slack in the expected way
func TestGettingStarted(t *testing.T) {
	client := newTestClient([]string{"ABC123"})
	testApp := client.CreateApp()

	go func() {
		err := testApp.Run(handler)
		if err != nil {
			t.Errorf("error running app: %v", err)
		}
	}()

	// Send hello message
	client.SendEventToApp(messageEvent(
		slackevents.MessageEvent{
			Text:    "hello",
			Channel: "ABC123",
			User:    "DEF456",
		},
	))

	// Expect a single message and clear the message list
	if len(client.messagesSent) != 1 {
		t.Errorf("expected one message to be sent, got %d", len(client.messagesSent))
	}
	msg := client.messagesSent[0]
	client.messagesSent = nil

	firstBlocks, _ := json.MarshalIndent(msg.blocks, "", "  ")
	if !strings.Contains(string(firstBlocks), `Hello to you too: `) {
		t.Errorf("first message content was not as expected, got: %v", string(firstBlocks))
	}

	// Select an option value
	client.SendEventToApp(messageInteractionEvent(
		"hash",
		"timestamp",
		msg.metadata,
		slack.ActionCallbacks{},
		&slack.BlockActionStates{
			Values: map[string]map[string]slack.BlockAction{
				fmt.Sprintf("input-0-%v", hashstr(strings.Join([]string{"a", "b", "c"}, ","))): {
					"x": slack.BlockAction{
						SelectedOption: slack.OptionBlockObject{
							Value: "c",
						},
					},
				},
			},
		},
	))

	if len(client.messagesSent) != 1 {
		t.Errorf("expected one message to be sent, got %d", len(client.messagesSent))
	}
	msg = client.messagesSent[0]
	client.messagesSent = nil

	secondBlocks, _ := json.MarshalIndent(msg.blocks, "", "  ")
	if !strings.Contains(string(secondBlocks), `You chose \"c\"`) {
		t.Errorf("message content was not as expected, got: %v", string(secondBlocks))
	}
}

// handler should be kept in sync with README.md and examples/gettingstarted/main.go
func handler(ctx context.Context, ev spanner.Event) {
	if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {
		reply := ev.SendMessage(msg.Channel().ID())
		reply.PlainText(fmt.Sprintf("Hello to you too: %v", msg.User()))

		letter := reply.Select("Pick a letter", spanner.Options("a", "b", "c"))
		if letter != "" {
			ev.SendMessage(msg.Channel().ID()).PlainText(fmt.Sprintf("You chose %q", letter))
		}
	}
}
