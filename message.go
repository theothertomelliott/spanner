package chatframework

import (
	"fmt"

	"github.com/slack-go/slack"
)

var _ ReceivedMessage = &receivedMessageSlack{}

type receivedMessageSlack struct {
	eventMetadataSlack

	TextInternal string `json:"text"`

	messages []*messageSlack

	ChildMessage *messageSlack `json:"child_message"`
}

func (m *receivedMessageSlack) Text() string {
	return m.TextInternal
}

func (m *receivedMessageSlack) SendMessage() Message {
	message := &messageSlack{
		blocksSlack: &blocksSlack{},
		ChannelID:   m.ChannelInternal,
		unsent:      true, // This means the message was created in this event loop
	}
	m.messages = append(m.messages, message)

	return message
}

func (m *receivedMessageSlack) handleRequest(req requestSlack) error {
	for _, message := range m.messages {
		m.ChildMessage = message

		err := message.handleRequest(req)
		if err != nil {
			return err
		}
	}

	var payload interface{} = map[string]interface{}{}
	req.client.Ack(req.req, payload)

	return nil
}

func (m *receivedMessageSlack) populateEvent(p eventPopulation) error {
	if m.ChildMessage != nil {
		return m.ChildMessage.populateEvent(p)
	}
	return nil
}

var _ Message = &messageSlack{}

type messageSlack struct {
	*blocksSlack

	ChannelID    string `json:"channel_id"`
	MessageIndex string `json:"message_index"`
	unsent       bool
}

func (m *messageSlack) Channel(channelID string) {
	panic("not implemented")
}

func (m *messageSlack) Text(message string) {
	m.addText(message)
}

func (m *messageSlack) Select(title string, options []string) string {
	_, _ = m.addSelect(title, options)

	// TODO: get the state

	return ""
}

func (m *messageSlack) handleRequest(req requestSlack) error {
	if m.unsent {
		_, _, _, err := req.client.SendMessage(
			m.ChannelID,
			slack.MsgOptionBlocks(m.blocks...),
			slack.MsgOptionMetadata(slack.SlackMetadata{
				EventType: "bot_message",
				EventPayload: map[string]interface{}{
					"metadata": string(req.Metadata()),
				},
			}))
		if err != nil {
			return fmt.Errorf("sending message: %w", err)
		}
	}

	return nil
}

func (m *messageSlack) populateEvent(p eventPopulation) error {

	return nil
}
