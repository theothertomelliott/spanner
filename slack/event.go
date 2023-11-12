package slack

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

var _ spanner.Event = &event{}

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

func (e *event) Connected() bool {
	return e.state.Connected
}

func (e *event) JoinChannel(channel string) {
	e.channelsToJoin = append(e.channelsToJoin, channel)
}

func (e *event) Custom() spanner.CustomEvent {
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

func (e *event) SlashCommand(command string) spanner.SlashCommand {
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

func (e *event) finishEvent(req request) error {
	for _, channel := range e.channelsToJoin {
		err := e.doJoinChannel(channel, req)
		if err != nil {
			return err
		}
	}

	err := e.state.sendMessages(req)
	if err != nil {
		return err
	}

	if message := e.state.Message; message != nil {
		return message.finishEvent(req)
	}

	if slashCommand := e.state.SlashCommand; slashCommand != nil {
		return slashCommand.finishEvent(req)
	}

	// Handle the event if nothing else does
	var payload interface{} = map[string]interface{}{}
	req.client.Ack(req.req, payload)

	return nil
}

var channelIDRegex = regexp.MustCompile("^[a-z0-9-]{1}[a-z0-9-]{0,20}$")

func (e *event) doJoinChannel(channel string, req request) error {
	if channelIDRegex.MatchString(channel) {
		_, _, _, err := req.client.JoinConversation(channel)
		if err != nil {
			return err
		}
	}

	// Remove any hashes at the start of the channel name
	channel = strings.TrimLeft(channel, "#")

	authTest, err := req.client.AuthTest()
	if err != nil {
		return err
	}

	channels, err := getAllConversations(req.client, authTest.UserID)
	if err != nil {
		return err
	}
	for _, c := range channels {
		if c.Name == channel || c.ID == channel {
			// Already in this channel
			fmt.Println("Already in the channel:", channel)
			return nil
		}
	}

	allChannels, err := getAllConversations(req.client, "")
	if err != nil {
		return err
	}

	for _, c := range allChannels {
		if c.Name == channel || c.ID == channel {
			_, _, _, err = req.client.JoinConversation(c.ID)
			if err != nil {
				return err
			}
			fmt.Println("Joined the channel:", channel)
			return nil
		}
	}

	return nil
}

// TODO: Short circuit pagination as needed
// TODO: Allow for caching of channel lists
func getAllConversations(client *socketmode.Client, userID string) ([]slack.Channel, error) {
	var (
		nextCursor      string = "more"
		cursor          string = ""
		channels        []slack.Channel
		currentChannels []slack.Channel
		err             error
	)
	for nextCursor != "" {
		if userID != "" {
			currentChannels, nextCursor, err = client.GetConversationsForUser(&slack.GetConversationsForUserParameters{
				UserID:          userID,
				Cursor:          cursor,
				Limit:           200,
				ExcludeArchived: true,
			})
		} else {
			currentChannels, nextCursor, err = client.GetConversations(&slack.GetConversationsParameters{
				Cursor:          cursor,
				Limit:           200,
				ExcludeArchived: true,
			})
		}
		if err != nil {
			return nil, err
		}
		cursor = nextCursor
		channels = append(channels, currentChannels...)
	}
	return channels, nil
}

type eventPopulation struct {
	interactionCallbackEvent slack.InteractionCallback
	interaction              slack.InteractionType
	messageIndex             string
}

func parseCombinedEvent(client *socketmode.Client, ce combinedEvent) *event {
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
			p := eventPopulation{
				interactionCallbackEvent: interactionCallbackEvent,
				interaction:              interactionCallbackEvent.Type,
				messageIndex:             messageIndex,
			}

			if out.state.MessageSender != nil {
				err := out.state.MessageSender.populateEvent(p)
				if err != nil {
					panic(err)
				}
			}
			if out.state.Message != nil {
				err := out.state.Message.populateEvent(p)
				if err != nil {
					panic(err)
				}
			}

		}

		return out
	}

	return out
}
