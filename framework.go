package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type App interface {
	Run(func(ev EventState) error) error
}

func NewSlackApp(botToken string, appToken string) (*slackApp, error) {
	if !strings.HasPrefix(botToken, "xoxb-") {
		return nil, fmt.Errorf("bot token must be the token with prefix 'xoxb-'")
	}
	if !strings.HasPrefix(appToken, "xapp-") {
		return nil, fmt.Errorf("app token must be the token with prefix 'xapp-'")
	}

	api := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	return &slackApp{
		api:    api,
		client: client,
	}, nil
}

type slackApp struct {
	api    *slack.Client
	client *socketmode.Client
}

func (s *slackApp) Run(handler func(ev EventState) error) error {
	go func() {
		for evt := range s.client.Events {
			es := parseSlackEvent(evt)
			err := handler(es)
			if err != nil {
				log.Printf("handling event: %v", err)
				continue // Move on without acknowledging, will force a repeat
			}
			if evt.Request != nil {
				if modal := es.interaction.renderModal(); modal != nil {
					_, err := s.client.OpenView(es.interaction.triggerID, *modal)
					if err != nil {
						log.Printf("handling event: %v", err)
						continue
					}
				}
				var payload interface{} = map[string]interface{}{}
				if es.interaction != nil {
					payload = es.interaction.payload()
				}
				s.client.Ack(*evt.Request, payload)
			}
		}
	}()
	return s.client.Run()
}

type EventState interface {
	ReceiveMessage() *Message
	SlashCommand(command string) Interaction
}

type Interaction interface {
	// TODO: Refactor so that interactions create a separate object for handling modals, messages, etc
	Modal(string)
	Text(string)
}

type Message struct {
	UserID string
	Text   string
}

type interactionSlack struct {
	triggerID string

	modal bool

	// modal only
	title string

	blocks []slack.Block
}

func (is *interactionSlack) Modal(title string) {
	is.modal = true
	is.title = title
}

func (is *interactionSlack) Text(message string) {
	is.blocks = append(is.blocks, slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: message,
		},
		nil,
		nil,
		// slack.NewAccessory(
		// 	slack.NewButtonBlockElement(
		// 		"",
		// 		"somevalue",
		// 		&slack.TextBlockObject{
		// 			Type: slack.PlainTextType,
		// 			Text: "bar",
		// 		},
		// 	),
		// ),
	))
}

func (is *interactionSlack) renderModal() *slack.ModalViewRequest {
	if is == nil {
		return nil
	}
	if !is.modal {
		return nil
	}
	modal := &slack.ModalViewRequest{
		Type:  slack.ViewType("modal"),
		Title: slack.NewTextBlockObject(slack.PlainTextType, is.title, false, false),
		Close: slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false),
		//Submit: slack.NewTextBlockObject(slack.PlainTextType, submitText, false, false),
		Blocks: slack.Blocks{
			BlockSet: is.blocks,
		},
	}
	return modal
}

func (is *interactionSlack) payload() interface{} {
	if is.modal {
		return nil
	}
	return map[string]interface{}{
		"blocks": is.blocks,
	}
}

func parseSlackEvent(ev socketmode.Event) *eventStateSlack {
	return &eventStateSlack{
		event: ev,
	}
}

type eventStateSlack struct {
	event socketmode.Event

	interaction   *interactionSlack
	messageOutbox []Message
}

func (e *eventStateSlack) ReceiveMessage() *Message {
	if e.event.Type != socketmode.EventTypeEventsAPI {
		return nil
	}
	eventsAPIEvent, ok := e.event.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return nil
	}
	if eventsAPIEvent.Type != slackevents.CallbackEvent {
		return nil
	}

	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		me := Message{
			UserID: ev.User,
			Text:   ev.Text,
		}
		return &me
	}
	return nil
}

func (e *eventStateSlack) SlashCommand(command string) Interaction {
	if e.event.Type != socketmode.EventTypeSlashCommand {
		return nil
	}
	cmd, ok := e.event.Data.(slack.SlashCommand)
	if !ok {
		return nil
	}

	if cmd.Command == command {
		e.interaction = &interactionSlack{
			triggerID: cmd.TriggerID,
		}
		return e.interaction
	}
	return nil
}
