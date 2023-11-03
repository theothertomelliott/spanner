package chatframework

import (
	"fmt"

	"github.com/slack-go/slack"
)

var _ ReceivedMessage = &receivedMessageSlack{}

type receivedMessageSlack struct {
	eventMetadataSlack
	*MessageSenderSlack `json:"ms"`

	TextInternal string `json:"text"`
}

func (m *receivedMessageSlack) handleRequest(req requestSlack) error {
	err := m.MessageSenderSlack.sendMessages(req)
	if err != nil {
		return err
	}

	var payload interface{} = map[string]interface{}{}
	req.client.Ack(req.req, payload)

	return nil
}

func (m *receivedMessageSlack) populateEvent(p eventPopulation) error {
	if m.MessageSenderSlack != nil {
		err := m.MessageSenderSlack.populateEvent(p)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ Message = &messageSlack{}

type messageSlack struct {
	*BlocksSlack `json:"blocks"` // This ensures that the value is not nil

	ChannelID    string `json:"channel_id"`
	MessageIndex string `json:"message_index"`
	unsent       bool
}

func (m *messageSlack) Channel(channelID string) {
	m.ChannelID = channelID
}

func (m *messageSlack) handleRequest(req requestSlack) error {
	if m.unsent {
		_, _, _, err := req.client.SendMessage(
			m.ChannelID,
			slack.MsgOptionBlocks(m.blocks...),
			slack.MsgOptionMetadata(slack.SlackMetadata{
				EventType: "bot_message",
				EventPayload: map[string]interface{}{
					"message_index": m.MessageIndex,
					"metadata":      string(req.Metadata()),
				},
			}))
		if err != nil {
			return fmt.Errorf("sending message: %w", err)
		}
	}

	return nil
}

func (m *messageSlack) populateEvent(p eventPopulation) error {
	m.BlockStates = blockActionToState(p.interactionCallbackEvent.BlockActionState.Values)
	return nil
}

type MessageSenderSlack struct {
	readMessageIndex int             // track offset of messages so we don't create extra when processing actions
	Messages         []*messageSlack `json:"messages"`

	defaultChannelID string
}

func (m *receivedMessageSlack) Text() string {
	return m.TextInternal
}

func (m *MessageSenderSlack) SendMessage() Message {
	if m.readMessageIndex < len(m.Messages) {
		return m.Messages[m.readMessageIndex]
	}
	message := &messageSlack{
		BlocksSlack:  &BlocksSlack{},
		MessageIndex: fmt.Sprintf("%v", len(m.Messages)),
		ChannelID:    m.defaultChannelID,
		unsent:       true, // This means the message was created in this event loop

	}
	m.Messages = append(m.Messages, message)
	m.readMessageIndex++

	return message
}

func (m *MessageSenderSlack) sendMessages(req requestSlack) error {
	for _, message := range m.Messages {
		err := message.handleRequest(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MessageSenderSlack) populateEvent(p eventPopulation) error {
	for _, message := range m.Messages {
		if message.MessageIndex == p.messageIndex {
			return message.populateEvent(p)
		}
	}
	return nil
}
