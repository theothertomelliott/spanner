package slack

import (
	"encoding/json"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/chatframework"
)

type event struct {
	hash string

	state eventState
}

type eventMetadata struct {
	ChannelInternal string `json:"channel"`
	UserInternal    string `json:"user"`
}

func (e eventMetadata) User() string {
	return e.UserInternal
}

func (e eventMetadata) Channel() string {
	return e.ChannelInternal
}

type eventState struct {
	Metadata     eventMetadata    `json:"metadata"`
	SlashCommand *slashCommand    `json:"slash_command"`
	Message      *receivedMessage `json:"message"`
}

func parseSlackEvent(ev socketmode.Event) *event {
	out := &event{}

	if ev.Type == socketmode.EventTypeSlashCommand {
		cmd, ok := ev.Data.(slack.SlashCommand)
		if !ok {
			return out
		}

		out.state.Metadata.ChannelInternal = cmd.ChannelID
		out.state.Metadata.UserInternal = cmd.UserID

		out.state.SlashCommand = &slashCommand{
			eventMetadata: out.state.Metadata,
			MessageSender: &MessageSender{
				DefaultChannelID: out.state.Metadata.ChannelInternal,
			},

			TriggerID: cmd.TriggerID,
			Command:   cmd.Command,
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
				out.state.Metadata.ChannelInternal = ev.Channel
				out.state.Metadata.UserInternal = ev.User

				out.state.Message = &receivedMessage{
					eventMetadata: out.state.Metadata,
					TextInternal:  ev.Text,
					MessageSender: &MessageSender{
						DefaultChannelID: out.state.Metadata.ChannelInternal,
					},
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
					eventPopulation{
						interactionCallbackEvent: interactionCallbackEvent,
						interaction:              interactionCallbackEvent.Type,
						messageIndex:             "",
					},
				)
			}

		} else if eventMeta := interactionCallbackEvent.Message.Metadata; eventMeta.EventType == "bot_message" {
			messageIndex := eventMeta.EventPayload["message_index"].(string)
			err := json.Unmarshal([]byte(eventMeta.EventPayload["metadata"].(string)), &out.state)
			if err != nil {
				panic(err)
			}
			if out.state.Message != nil {
				out.state.Message.populateEvent(
					eventPopulation{
						interactionCallbackEvent: interactionCallbackEvent,
						interaction:              interactionCallbackEvent.Type,
						messageIndex:             messageIndex,
					},
				)
			}

		} else {
			fmt.Println("no metadata")
		}
	}
	return out
}

func (e *event) ReceiveMessage() chatframework.ReceivedMessage {
	if e.state.Message != nil {
		return e.state.Message
	}
	return nil
}

func (e *event) SlashCommand(command string) chatframework.SlashCommand {
	if e.state.SlashCommand == nil {
		return nil
	}
	if e.state.SlashCommand.Command != command {
		return nil
	}
	return e.state.SlashCommand
}

type eventPopulation struct {
	interactionCallbackEvent slack.InteractionCallback
	interaction              slack.InteractionType
	messageIndex             string
}
