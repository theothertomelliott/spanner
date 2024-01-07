package slack

import (
	"context"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

func newTestClient() *testClient {
	return &testClient{
		Events:    make(chan socketmode.Event),
		stop:      make(chan struct{}),
		postEvent: make(chan interface{}, 10),
	}
}

type testClient struct {
	nilSocketClient

	messagesSent    []sentMessage
	messagesUpdated []updatedMessage

	Events chan socketmode.Event

	postEvent chan interface{}

	stop chan struct{}

	runMtx   sync.Mutex
	runCount int
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

// CreateApp returns a spanner.App that uses this client
func (r *testClient) CreateApp() spanner.App {
	testApp := newAppWithClient(
		r,
		r.Events,
	)

	testApp.SetPostEventFunc(r.PostEventFunc)

	return testApp
}

// PostEventFunc provides a spanner.PostEventFunc to use with a test app
// This is automatically applied by the CreateApp function
func (r *testClient) PostEventFunc(ctx context.Context) {
	r.postEvent <- struct{}{}
}

// SendEventToApp sends an event to the connected app (created with CreateApp)
// and blocks until the event is handled.
func (r *testClient) SendEventToApp(e socketmode.Event) {
	r.Events <- e
	<-r.postEvent
}

// SendEventToAppAsync sends an event without blocking.
func (r *testClient) SendEventToAppAsync(e socketmode.Event) {
	go func() {
		r.Events <- e
	}()
}

func (r *testClient) RunContext(context.Context) error {
	r.runMtx.Lock()
	r.runCount++
	r.runMtx.Unlock()
	<-r.stop // Must block
	return nil
}

func (*testClient) Ack(req socketmode.Request, payload ...interface{}) {}

func (c *testClient) SendMessageWithMetadata(ctx context.Context, channelID string, blocks []slack.Block, metadata slack.SlackMetadata) (string, string, string, error) {
	c.messagesSent = append(c.messagesSent, sentMessage{
		channelID: channelID,
		blocks:    blocks,
		metadata:  metadata,
	})
	return "", "", "", nil
}

func (c *testClient) UpdateMessageWithMetadata(ctx context.Context, channelID string, timestamp string, blocks []slack.Block, metadata slack.SlackMetadata) (string, string, string, error) {
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

func messageEvent(messageEvent slackevents.MessageEvent) socketmode.Event {
	return socketmode.Event{
		Type: socketmode.EventTypeEventsAPI,
		Data: slackevents.EventsAPIEvent{
			Type: slackevents.CallbackEvent,
			InnerEvent: slackevents.EventsAPIInnerEvent{
				Data: &messageEvent,
			},
		},
	}
}

func slashCommandEvent(data slack.SlashCommand) socketmode.Event {
	return socketmode.Event{
		Type: socketmode.EventTypeSlashCommand,
		Data: data,
	}
}

func messageInteractionEvent(
	hash string,
	timestamp string,
	metadata slack.SlackMetadata,
	actionCallbacks slack.ActionCallbacks,
	blockActionState *slack.BlockActionStates,
) socketmode.Event {
	return socketmode.Event{
		Type: socketmode.EventTypeInteractive,
		Data: slack.InteractionCallback{
			ViewSubmissionCallback: slack.ViewSubmissionCallback{
				Hash: hash,
			},
			Message: slack.Message{
				Msg: slack.Msg{
					Metadata:  metadata,
					Timestamp: timestamp,
				},
			},
			ActionCallback:   actionCallbacks,
			BlockActionState: blockActionState,
		},
	}
}
