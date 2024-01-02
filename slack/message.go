package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/theothertomelliott/spanner"
)

var _ spanner.ReceivedMessage = &receivedMessage{}

type receivedMessage struct {
	eventMetadata

	TextInternal string `json:"text"`
}

func (m *receivedMessage) finishEvent(req request) error {
	// Placeholder for actions specific to received messages

	var payload interface{} = map[string]interface{}{}
	req.client.Ack(req.req, payload)

	return nil
}

func (m *receivedMessage) populateEvent(p eventPopulation) error {
	// Placeholder for actions specific to received messages
	return nil
}

var _ spanner.Message = &message{}

type message struct {
	*Blocks `json:"blocks"` // This ensures that the value is not nil

	ChannelID         string `json:"channel_id"`
	MessageIndex      string `json:"message_index"`
	PreviousBlockHash string `json:"previous_block_hash"`
	MessageTS         string `json:"message_ts"`
	actionMessageTS   string `json:"-"`
}

func (m *message) Channel(channelID string) {
	m.ChannelID = channelID
}

func (m *message) finishEvent(req request) error {
	if m.MessageTS == "" {
		_, timestamp, _, err := req.client.SendMessageWithMetadata(
			context.TODO(),
			m.ChannelID,
			m.blocks,
			slack.SlackMetadata{
				EventType: "bot_message",
				EventPayload: map[string]interface{}{
					"message_index": m.MessageIndex,
					"metadata":      string(req.Metadata()),
				},
			})
		if err != nil {
			return fmt.Errorf("sending message: %w", renderSlackError(err))
		}

		// Record the timestamp in the message metadata
		// This will allow us to match messages to the correct message in the hierarchy
		m.MessageTS = timestamp
		_, _, _, err = req.client.UpdateMessageWithMetadata(
			context.TODO(),
			m.ChannelID,
			m.MessageTS,
			m.blocks,
			slack.SlackMetadata{
				EventType: "bot_message",
				EventPayload: map[string]interface{}{
					"message_index": m.MessageIndex,
					"metadata":      string(req.Metadata()),
				},
			})
		if err != nil {
			return fmt.Errorf("updating message to record timestamp: %w", renderSlackError(err))
		}
	} else if m.MessageTS != m.actionMessageTS {
		_, _, _, err := req.client.UpdateMessageWithMetadata(
			context.TODO(),
			m.ChannelID,
			m.MessageTS,
			m.blocks,
			slack.SlackMetadata{
				EventType: "bot_message",
				EventPayload: map[string]interface{}{
					"message_index": m.MessageIndex,
					"metadata":      string(req.Metadata()),
				},
			})
		if err != nil {
			return fmt.Errorf("updating message: %w", renderSlackError(err))
		}
	}

	return nil
}

func (m *message) populateEvent(p eventPopulation) error {
	m.BlockStates = blockActionToState(p)
	m.actionMessageTS = p.interactionCallbackEvent.Message.Timestamp
	return nil
}

type MessageSender struct {
	readMessageIndex int        // track offset of messages so we don't create extra when processing actions
	Messages         []*message `json:"messages"`
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
		ChannelID:    channelID,
	}
	m.Messages = append(m.Messages, message)

	return message
}

func (m *MessageSender) sendMessages(req request) error {
	for _, message := range m.Messages {
		err := message.finishEvent(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MessageSender) populateEvent(p eventPopulation) error {
	for _, message := range m.Messages {
		if message.MessageIndex == p.messageIndex {
			return message.populateEvent(p)
		}
	}
	return nil
}
