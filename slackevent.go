package chatframework

import (
	"encoding/json"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type eventSlack struct {
	hash string

	state eventState
}

type eventMetadataSlack struct {
	ChannelInternal string `json:"channel"`
	UserInternal    string `json:"user"`
}

func (e eventMetadataSlack) User() string {
	return e.UserInternal
}

func (e eventMetadataSlack) Channel() string {
	return e.ChannelInternal
}

type eventState struct {
	Metadata     eventMetadataSlack    `json:"metadata"`
	SlashCommand *slashCommandSlack    `json:"slash_command"`
	Message      *receivedMessageSlack `json:"message"`
}

func parseSlackEvent(ev socketmode.Event) *eventSlack {
	out := &eventSlack{}

	if ev.Type == socketmode.EventTypeSlashCommand {
		cmd, ok := ev.Data.(slack.SlashCommand)
		if !ok {
			return out
		}

		out.state.Metadata.ChannelInternal = cmd.ChannelID
		out.state.Metadata.UserInternal = cmd.UserID

		out.state.SlashCommand = &slashCommandSlack{
			eventMetadataSlack: out.state.Metadata,
			TriggerID:          cmd.TriggerID,
			Command:            cmd.Command,
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

				out.state.Message = &receivedMessageSlack{
					eventMetadataSlack: out.state.Metadata,
					TextInternal:       ev.Text,
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

func (e *eventSlack) ReceiveMessage() ReceivedMessage {
	if e.state.Message != nil {
		return e.state.Message
	}
	return nil
}

func (e *eventSlack) SlashCommand(command string) SlashCommand {
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
