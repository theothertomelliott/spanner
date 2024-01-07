package slack

import (
	"context"
	"encoding/json"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

type eventPopulator interface {
	populateEvent(ctx context.Context, p eventPopulation, depth int) error
}

type eventFinisher interface {
	finishEvent(ctx context.Context, req request) error
}

var _ spanner.Event = &event{}
var _ eventFinisher = &event{}

type event struct {
	hash string

	state eventState

	channelsToJoin []string
}

type eventMetadata struct {
	ChannelInfo *channel `json:"channel_info"`
	UserInfo    *user    `json:"user_info"`
}

func (e eventMetadata) User() spanner.User {
	return e.UserInfo
}

func (e eventMetadata) Channel() spanner.Channel {
	return e.ChannelInfo
}

type eventState struct {
	*MessageSender `json:"ms"`

	Metadata     eventMetadata    `json:"metadata"`
	Connected    bool             `json:"connected"`
	SlashCommand *slashCommand    `json:"slash_command"`
	Message      *receivedMessage `json:"message"`
	Custom       *customEvent     `json:"customEvent"`
}

func (e *event) ReceiveConnected() bool {
	return e.state.Connected
}

func (e *event) JoinChannel(channelID string) {
	e.channelsToJoin = append(e.channelsToJoin, channelID)
}

func (e *event) ReceiveCustomEvent() spanner.CustomEvent {
	if e.state.Custom != nil {
		return e.state.Custom
	}
	return nil
}

func (e *event) ReceiveMessage() spanner.ReceivedMessage {
	if e.state.Message != nil {
		return e.state.Message
	}
	return nil
}

func (e *event) ReceiveSlashCommand(command string) spanner.SlashCommand {
	if e.state.SlashCommand == nil {
		return nil
	}
	if e.state.SlashCommand.Command != command {
		return nil
	}
	return e.state.SlashCommand
}

func (e *event) SendMessage(channelID string) spanner.Message {
	return e.state.SendMessage(channelID)
}

func (e *event) finishEvent(ctx context.Context, req request) error {
	for _, channel := range e.channelsToJoin {
		err := e.doJoinChannel(ctx, channel, req)
		if err != nil {
			return err
		}
	}

	err := e.state.finishEvent(ctx, req)
	if err != nil {
		return err
	}

	if message := e.state.Message; message != nil {
		return message.finishEvent(ctx, req)
	}

	if slashCommand := e.state.SlashCommand; slashCommand != nil {
		return slashCommand.finishEvent(ctx, req)
	}

	// Handle the event if nothing else does
	var payload interface{} = map[string]interface{}{}
	req.client.Ack(req.req, payload)

	return nil
}

func (e *event) doJoinChannel(ctx context.Context, channel string, req request) error {
	_, _, _, err := req.client.JoinConversationContext(ctx, channel)
	if err != nil {
		return err
	}
	return nil
}

type eventPopulation struct {
	interactionCallbackEvent slack.InteractionCallback
	interaction              slack.InteractionType
	messageIndex             string
	interactionDepth         int
}

func parseCombinedEvent(ctx context.Context, client socketClient, ce combinedEvent) *event {
	out := &event{
		state: eventState{
			MessageSender: &MessageSender{},
		},
	}

	defer func() {
		// Set clients in metadata
		if out.state.Metadata.ChannelInfo != nil {
			out.state.Metadata.ChannelInfo.client = client
		}
		if out.state.Metadata.UserInfo != nil {
			out.state.Metadata.UserInfo.client = client
		}

		// TODO: There's probably a better way to do this
		if out.state.SlashCommand != nil {
			if out.state.SlashCommand.eventMetadata.ChannelInfo != nil {
				out.state.SlashCommand.eventMetadata.ChannelInfo.client = client
			}
			if out.state.SlashCommand.eventMetadata.UserInfo != nil {
				out.state.SlashCommand.eventMetadata.UserInfo.client = client
			}
		}
		if out.state.Message != nil {
			if out.state.Message.eventMetadata.ChannelInfo != nil {
				out.state.Message.eventMetadata.ChannelInfo.client = client
			}
			if out.state.Message.eventMetadata.UserInfo != nil {
				out.state.Message.eventMetadata.UserInfo.client = client
			}
		}
	}()

	if ce.customEvent != nil {
		out.state.Custom = ce.customEvent
		return out
	}

	var ev socketmode.Event
	if ce.ev != nil {
		ev = *ce.ev
	} else {
		return out
	}

	if ev.Type == socketmode.EventTypeConnected {
		out.state.Connected = true
		return out
	}

	if ev.Type == socketmode.EventTypeSlashCommand {
		cmd, ok := ev.Data.(slack.SlashCommand)
		if !ok {
			return out
		}

		out.state.Metadata.ChannelInfo = &channel{
			client:       client,
			Loaded:       true,
			IDInternal:   cmd.ChannelID,
			NameInternal: cmd.ChannelName,
		}
		out.state.Metadata.UserInfo = &user{
			client:     client,
			IDInternal: cmd.UserID,
		}

		out.state.SlashCommand = &slashCommand{
			eventMetadata: out.state.Metadata,
			TriggerID:     cmd.TriggerID,
			Command:       cmd.Command,
		}
		return out
	}

	if ev.Type == socketmode.EventTypeEventsAPI {
		eventsAPIEvent, ok := ev.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return out
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.MessageEvent:
				out.state.Metadata.ChannelInfo = &channel{
					client:     client,
					IDInternal: ev.Channel,
				}
				out.state.Metadata.UserInfo = &user{
					client:       client,
					IDInternal:   ev.User,
					NameInternal: ev.Username,
				}

				out.state.Message = &receivedMessage{
					eventMetadata: out.state.Metadata,
					TextInternal:  ev.Text,
				}
			}
			return out
		}
		return out
	}

	if ev.Type == socketmode.EventTypeInteractive {
		interactionCallbackEvent, ok := ev.Data.(slack.InteractionCallback)
		if !ok {
			return out
		}

		out.hash = interactionCallbackEvent.Hash

		if metadata := interactionCallbackEvent.View.PrivateMetadata; metadata != "" {
			err := json.Unmarshal([]byte(metadata), &out.state)
			if err != nil {
				panic(err)
			}
			if out.state.SlashCommand != nil {
				out.state.SlashCommand.populateEvent(
					ctx,
					eventPopulation{
						interactionCallbackEvent: interactionCallbackEvent,
						interaction:              interactionCallbackEvent.Type,
						messageIndex:             "",
					},
					0,
				)
			}

		} else if eventMeta := interactionCallbackEvent.Message.Metadata; eventMeta.EventType == "bot_message" {
			messageIndex := eventMeta.EventPayload["message_index"].(string)
			err := json.Unmarshal([]byte(eventMeta.EventPayload["metadata"].(string)), &out.state)
			if err != nil {
				panic(err)
			}
			p := eventPopulation{
				interactionCallbackEvent: interactionCallbackEvent,
				interaction:              interactionCallbackEvent.Type,
				messageIndex:             messageIndex,
			}

			if out.state.MessageSender != nil {
				err := out.state.MessageSender.populateEvent(ctx, p, 0)
				if err != nil {
					panic(err)
				}
			}
			if out.state.Message != nil {
				err := out.state.Message.populateEvent(ctx, p, 0)
				if err != nil {
					panic(err)
				}
			}

		}

		return out
	}

	return out
}
