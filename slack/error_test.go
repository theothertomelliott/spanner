package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/slack-go/slack/slackevents"
	"github.com/theothertomelliott/spanner"
)

// TestErrorHandling verifies that error handlers are called appropriately
func TestErrorHandling(t *testing.T) {
	client := newTestClient([]string{"ABC123"})
	testApp := client.CreateApp()

	go func() {
		err := testApp.Run(handlerTestErrors)
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
	if len(client.messagesSent) != 2 {
		t.Errorf("expected two messages to be sent, got %d", len(client.messagesSent))
	}

	firstBlocks, _ := json.MarshalIndent(client.messagesSent[0].blocks, "", "  ")
	if !strings.Contains(string(firstBlocks), `This message should succeed`) {
		t.Errorf("first message content was not as expected, got: %v", string(firstBlocks))
	}

	secondBlocks, _ := json.MarshalIndent(client.messagesSent[1].blocks, "", "  ")
	if !strings.Contains(string(secondBlocks), `There was an error sending a message`) {
		t.Errorf("first message content was not as expected, got: %v", string(secondBlocks))
	}
}

func handlerTestErrors(ctx context.Context, ev spanner.Event) error {
	if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {
		replyGood := ev.SendMessage(msg.Channel().ID())
		replyGood.PlainText("This message should succeed")
		replyGood.ErrorFunc(func(ctx context.Context, ev spanner.ErrorEvent) {
			panic("did not expect this message to fail")
		})

		replyBad := ev.SendMessage("invalid_channel")
		replyBad.PlainText("This message will always fail to post")
		replyBad.ErrorFunc(func(ctx context.Context, ev spanner.ErrorEvent) {
			errorNotice := ev.SendMessage(msg.Channel().ID())
			errorNotice.PlainText(fmt.Sprintf("There was an error sending a message: %v", ev.ReceiveError()))
		})

		replySkipped := ev.SendMessage(msg.Channel().ID())
		replySkipped.PlainText("This message should be skipped because of the previous error")
		replySkipped.ErrorFunc(func(ctx context.Context, ev spanner.ErrorEvent) {
			panic("did not expect this message to fail")
		})
	}
	return nil
}
