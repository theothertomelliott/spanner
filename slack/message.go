package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/theothertomelliott/spanner"
)

var _ spanner.ReceivedMessage = &receivedMessage{}
var _ eventPopulator = &receivedMessage{}

type receivedMessage struct {
	eventMetadata

	TextInternal string `json:"text"`
}

func (m *receivedMessage) populateEvent(ctx context.Context, p eventPopulation, depth int) error {
	// Placeholder for actions specific to received messages
	return nil
}

var _ spanner.Message = &message{}
var _ eventPopulator = &message{}
var _ action = &message{}

type message struct {
	*Blocks `json:"blocks"` // This ensures that the value is not nil

	ChannelID           string `json:"channel_id"`
	MessageIndex        string `json:"message_index"`
	EventDepth          int    `json:"event_depth"`
	currentMessageIndex string
	currentEventDepth   int
	actionMessageTS     string
	unsent              bool
}

func (m *message) Type() string {
	return "message"
}

func (m *message) Data() interface{} {
	panic("unimplemented")
}

func (m *message) Channel(channelID string) {
	m.ChannelID = channelID
}

func (m *message) exec(ctx context.Context, req request) (interface{}, error) {
	if m.unsent {
		_, _, _, err := req.client.SendMessageWithMetadata(
			ctx,
			m.ChannelID,
			m.blocks,
			slack.SlackMetadata{
				EventType: "bot_message",
				EventPayload: map[string]interface{}{
					"message_index": m.MessageIndex,
					"event_depth":   m.EventDepth,
					"metadata":      string(req.Metadata()),
				},
			})
		if err != nil {
			return nil, fmt.Errorf("sending message: %w", renderSlackError(err))
		}

	} else if m.MessageIndex == m.currentMessageIndex && m.EventDepth == m.currentEventDepth {
		_, _, _, err := req.client.UpdateMessageWithMetadata(
			ctx,
			m.ChannelID,
			m.actionMessageTS,
			m.blocks,
			slack.SlackMetadata{
				EventType: "bot_message",
				EventPayload: map[string]interface{}{
					"message_index": m.MessageIndex,
					"event_depth":   m.EventDepth,
					"metadata":      string(req.Metadata()),
				},
			})
		if err != nil {
			return nil, fmt.Errorf("updating message: %w", renderSlackError(err))
		}
	}

	return nil, nil
}

func (m *message) populateEvent(ctx context.Context, p eventPopulation, depth int) error {
	m.BlockStates = blockActionToState(p)
	m.actionMessageTS = p.interactionCallbackEvent.Message.Timestamp
	m.currentEventDepth = p.interactionDepth
	m.currentMessageIndex = p.messageIndex
	return nil
}

var _ eventPopulator = &MessageSender{}

type MessageSender struct {
	actionQueue *actionQueue

	readMessageIndex int        // track offset of messages so we don't create extra when processing actions
	Messages         []*message `json:"messages"`
	EventDepth       int        `json:"event_depth"`
}

func (m *receivedMessage) Text() string {
	return m.TextInternal
}

func (m *MessageSender) SendMessage(channelID string) spanner.Message {
	defer func() {
		m.readMessageIndex++
	}()

	if m.readMessageIndex < len(m.Messages) {
		return m.Messages[m.readMessageIndex]
	}

	message := &message{
		Blocks:       &Blocks{},
		MessageIndex: fmt.Sprintf("%v", len(m.Messages)),
		EventDepth:   m.EventDepth,
		ChannelID:    channelID,
		unsent:       true,
	}
	m.Messages = append(m.Messages, message)

	m.actionQueue.enqueue(message)

	return message
}

func (m *MessageSender) populateEvent(ctx context.Context, p eventPopulation, depth int) error {
	m.EventDepth = depth
	for _, message := range m.Messages {
		if message.MessageIndex == p.messageIndex {
			m.actionQueue = p.actionQueue
			p.actionQueue.enqueue(message)
			return message.populateEvent(ctx, p, depth+1)
		}
	}
	return nil
}
