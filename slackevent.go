package chatframework

import (
	"encoding/json"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type eventSlack struct {
	hash string

	state eventState
}

type eventState struct {
	SlashCommand *slashCommandSlack `json:"slash_command"`
	Message      *Message           `json:"message"`
}

func parseSlackEvent(ev socketmode.Event) *eventSlack {
	out := &eventSlack{}

	if ev.Type == socketmode.EventTypeSlashCommand {
		cmd, ok := ev.Data.(slack.SlashCommand)
		if !ok {
			return out
		}

		out.state.SlashCommand = &slashCommandSlack{
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
				out.state.Message = &Message{
					UserID: ev.User,
					Text:   ev.Text,
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
				out.state.SlashCommand.ModalInternal.ReceivedView = &interactionCallbackEvent.View
				if interactionCallbackEvent.Type == slack.InteractionTypeBlockActions {
					out.state.SlashCommand.ModalInternal.update = action
				}
				if interactionCallbackEvent.Type == slack.InteractionTypeViewSubmission {
					out.state.SlashCommand.ModalInternal.update = submitted
				}
				if interactionCallbackEvent.Type == slack.InteractionTypeViewClosed {
					out.state.SlashCommand.ModalInternal.update = closed
				}
			}
		}
	}
	return out
}

func (e *eventSlack) ReceiveMessage() *Message {
	return e.state.Message
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
