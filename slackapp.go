package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

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
				if modal := es.modal.render(); modal != nil {
					_, err := s.client.OpenView(es.modal.triggerID, *modal)
					if err != nil {
						modalContent, jsonErr := json.MarshalIndent(modal, "", "  ")
						if jsonErr != nil {
							log.Println(jsonErr)
						}
						log.Printf("opening view: %v\n%v\n", err, string(modalContent))
						continue
					}
				}
				var payload interface{} = map[string]interface{}{}
				// TODO: Implement a response
				// if es.interaction != nil {
				// 	payload = es.interaction.payload()
				// }
				s.client.Ack(*evt.Request, payload)
			}
		}
	}()
	return s.client.Run()
}

type interactionSlack struct {
	triggerID string

	parent *eventStateSlack
}

func (is *interactionSlack) Modal(title string) Modal {
	is.parent.modal = &modalSlack{
		triggerID: is.triggerID,
		title:     title,
	}
	return is.parent.modal
}

type modalSlack struct {
	triggerID string

	// modal only
	title string

	inputID int

	blocks []slack.Block
}

func (m *modalSlack) Text(message string) {
	m.blocks = append(m.blocks, slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: message,
		},
		nil,
		nil,
	))
}

func (m *modalSlack) Select(text string, options []string) string {
	var optionObjects []*slack.OptionBlockObject
	for index, option := range options {
		optionObjects = append(
			optionObjects,
			slack.NewOptionBlockObject(
				fmt.Sprintf("input%voption%v", m.inputID, index),
				slack.NewTextBlockObject(slack.PlainTextType, option, false, false),
				nil,
			),
		)
	}

	input := slack.NewInputBlock(
		fmt.Sprintf("input%v", m.inputID),
		slack.NewTextBlockObject(slack.PlainTextType, text, false, false),
		nil,
		slack.NewOptionsSelectBlockElement(
			slack.OptTypeStatic,
			slack.NewTextBlockObject(slack.PlainTextType, text, false, false),
			fmt.Sprintf("input%vselection", m.inputID),
			optionObjects...,
		),
	)
	input.DispatchAction = true

	m.blocks = append(m.blocks,
		input,
	)
	m.inputID++

	// TODO: Empty options may not render
	if len(options) > 0 {
		return options[0]
	}
	return ""
}

func (m *modalSlack) Submit(text string) bool {
	panic("not imlemented")
}

func (m *modalSlack) render() *slack.ModalViewRequest {
	if m == nil {
		return nil
	}
	modal := &slack.ModalViewRequest{
		Type:  slack.ViewType("modal"),
		Title: slack.NewTextBlockObject(slack.PlainTextType, m.title, false, false),
		Close: slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false),

		// TODO: Should be controlled by the submit option
		// It should error out with a meaningful message if there are inputs but no submit button
		Submit: slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false),

		Blocks: slack.Blocks{
			BlockSet: m.blocks,
		},

		CallbackID: "slackFrameworkModal1", // TODO: Change this
	}
	return modal
}

func parseSlackEvent(ev socketmode.Event) *eventStateSlack {
	fmt.Printf("Event Type: %v\n", ev.Type)
	return &eventStateSlack{
		event: ev,
	}
}

type eventStateSlack struct {
	event socketmode.Event

	interaction *interactionSlack
	modal       *modalSlack
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

func (e *eventStateSlack) SlashCommand(command string) SlashCommand {
	if e.event.Type != socketmode.EventTypeSlashCommand {
		return nil
	}
	cmd, ok := e.event.Data.(slack.SlashCommand)
	if !ok {
		return nil
	}

	if cmd.Command == command {
		e.interaction = &interactionSlack{
			parent:    e,
			triggerID: cmd.TriggerID,
		}
		return e.interaction
	}
	return nil
}
