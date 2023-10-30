package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
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
		slack.OptionDebug(false),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(false),
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
				var payload interface{} = map[string]interface{}{}

				if slashCommand := es.state.SlashCommand; slashCommand != nil {
					if modal := slashCommand.ModalInternal.render(); modal != nil {
						metadata, err := json.MarshalIndent(es.state, "", "  ")
						if err != nil {
							log.Printf("saving metadata: %v\n", err)
						}
						modal.PrivateMetadata = string(metadata)

						if !es.isAction {
							_, err = s.client.OpenView(slashCommand.TriggerID, *modal)
							if err != nil {
								log.Printf("opening view: %v\n", err)
								continue
							}
						} else {
							_, err := s.client.UpdateView(
								*modal,
								slashCommand.ModalInternal.ReceivedView.ExternalID,
								es.hash,
								slashCommand.ModalInternal.ReceivedView.ID,
							)
							if err != nil {
								log.Printf("updating view: %v", err)
								continue
							}

							// TODO: Do this for submission events if we want to create a new view
							// Maybe we want to use push as well
							//payload = slack.NewUpdateViewSubmissionResponse(modal)
						}
					}
				}

				s.client.Ack(*evt.Request, payload)
			}
		}
	}()
	return s.client.Run()
}
