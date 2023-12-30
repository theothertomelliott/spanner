package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

// TestGettingStarted verifies that the code in examples/gettingstarted
// interacts with Slack in the expected way
func TestGettingStarted(t *testing.T) {
	slackEvents := make(chan socketmode.Event)
	client := newExampleTestClient()

	testApp := newAppWithClient(
		client,
		slackEvents,
	)

	go func() {
		err := testApp.Run(handler)
		if err != nil {
			t.Errorf("error running app: %v", err)
		}
	}()

	// Send hello message
	slackEvents <- messageEvent(
		slackevents.MessageEvent{
			Text:    "hello",
			Channel: "ABC123",
			User:    "DEF456",
		},
	)

	firstMetadata, firstBlocks, err := expectOneMessage(client.messagesSent, "ABC123")
	if err != nil {
		t.Errorf("receiving first message: %v", err)
	}
	if !strings.Contains(firstBlocks, `Hello to you too: `) {
		t.Errorf("first message content was not as expected, got: %v", string(firstBlocks))
	}

	slackEvents <- messageInteractionEvent(
		"hash",
		firstMetadata,
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
	)

	_, secondBlocks, err := expectOneMessage(client.messagesSent, "ABC123")
	if err != nil {
		t.Errorf("receiving second message: %v", err)
	}
	if !strings.Contains(secondBlocks, `You chose \"c\"`) {
		t.Errorf("message content was not as expected, got: %v", secondBlocks)
	}
}

// expectOneMessage checks for a single message being sent on the expected channel
// it returns the metadata and the JSON form of the message's blocks.
func expectOneMessage(messages chan sentMessage, channelID string) (slack.SlackMetadata, string, error) {
	var message sentMessage
	select {
	case message = <-messages:
	case <-time.After(time.Second):
		return slack.SlackMetadata{}, "", fmt.Errorf("timed out waiting for expected message")
	}

	// Ensure only one message was sent
	select {
	case <-messages:
		return slack.SlackMetadata{}, "", fmt.Errorf("expected exactly one message")
	case <-time.After(time.Second / 100):
	}

	blockJson, err := json.MarshalIndent(message.blocks, "", "  ")
	if err != nil {
		return slack.SlackMetadata{}, "", fmt.Errorf("could not marshal block data: %v", err)
	}

	return message.metadata, string(blockJson), nil
}

// handler should be kept in sync with README.md and examples/gettingstarted/main.go
func handler(ev spanner.Event) error {
	if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {
		reply := ev.SendMessage(msg.Channel().ID())
		reply.PlainText(fmt.Sprintf("Hello to you too: %v", msg.User()))

		letter := reply.Select("Pick a letter", spanner.Options("a", "b", "c"))
		if letter != "" {
			ev.SendMessage(msg.Channel().ID()).PlainText(fmt.Sprintf("You chose %q", letter))
		}
	}
	return nil
}

func newExampleTestClient() *exampleTestClient {
	return &exampleTestClient{
		messagesSent: make(chan sentMessage),
		stop:         make(chan struct{}),
	}
}

type exampleTestClient struct {
	nilSocketClient

	messagesSent chan sentMessage

	stop chan struct{}
}

type sentMessage struct {
	channelID string
	blocks    []slack.Block
	metadata  slack.SlackMetadata
}

func (r *exampleTestClient) RunContext(context.Context) error {
	<-r.stop // Must block
	return nil
}

func (*exampleTestClient) Ack(req socketmode.Request, payload ...interface{}) {}

func (c *exampleTestClient) SendMessageWithMetadata(ctx context.Context, channelID string, blocks []slack.Block, metadata slack.SlackMetadata) (string, string, string, error) {
	c.messagesSent <- sentMessage{
		channelID: channelID,
		blocks:    blocks,
		metadata:  metadata,
	}
	return "", "", "", nil
}