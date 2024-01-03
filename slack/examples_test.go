package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

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

	postEvent := make(chan interface{}, 10)
	testApp.SetPostEventFunc(func() {
		postEvent <- struct{}{}
	})

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

	// Wait for event to be handled
	<-postEvent

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
	slackEvents <- messageInteractionEvent(
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
	)

	// Wait for event to be handled
	<-postEvent

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
		stop: make(chan struct{}),
	}
}

type exampleTestClient struct {
	nilSocketClient

	messagesSent    []sentMessage
	messagesUpdated []updatedMessage

	stop chan struct{}
}

type sentMessage struct {
	channelID string
	blocks    []slack.Block
	metadata  slack.SlackMetadata
}

type updatedMessage struct {
	sentMessage
	timestamp string
}

func (r *exampleTestClient) RunContext(context.Context) error {
	<-r.stop // Must block
	return nil
}

func (*exampleTestClient) Ack(req socketmode.Request, payload ...interface{}) {}

func (c *exampleTestClient) SendMessageWithMetadata(ctx context.Context, channelID string, blocks []slack.Block, metadata slack.SlackMetadata) (string, string, string, error) {
	c.messagesSent = append(c.messagesSent, sentMessage{
		channelID: channelID,
		blocks:    blocks,
		metadata:  metadata,
	})
	return "", "", "", nil
}

func (c *exampleTestClient) UpdateMessageWithMetadata(ctx context.Context, channelID string, timestamp string, blocks []slack.Block, metadata slack.SlackMetadata) (string, string, string, error) {
	c.messagesUpdated = append(c.messagesUpdated, updatedMessage{
		sentMessage: sentMessage{
			channelID: channelID,
			blocks:    blocks,
			metadata:  metadata,
		},
		timestamp: timestamp,
	})
	return "", timestamp, "", nil
}
