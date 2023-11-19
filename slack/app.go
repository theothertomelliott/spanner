package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/spanner"
)

type AppConfig struct {
	BotToken string
	AppToken string
	Debug    bool
}

// NewApp creates a new slack app.
//
// botToken is the token for the bot user, with prefix 'xoxb-'
// appToken is the token for the app, with prefix 'xapp-'
//
// Slack apps use socket mode to handle events, so the app will need to be configured to use socket mode.
// https://api.slack.com/apis/connections/socket
//
// As at November 2023, this means that these apps cannot be distributed in the public Slack app directory.
func NewApp(config AppConfig) (spanner.App, error) {
	if !strings.HasPrefix(config.BotToken, "xoxb-") {
		return nil, fmt.Errorf("bot token must be the token with prefix 'xoxb-'")
	}
	if !strings.HasPrefix(config.AppToken, "xapp-") {
		return nil, fmt.Errorf("app token must be the token with prefix 'xapp-'")
	}

	api := slack.New(
		config.BotToken,
		slack.OptionDebug(config.Debug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(config.AppToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(config.Debug),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)
	events := client.Events

	return &app{
		client:        client,
		events:        events,
		combinedEvent: make(chan combinedEvent, 2),
		custom:        make(chan *customEvent, 2),
	}, nil
}

type app struct {
	client socketClient
	events chan socketmode.Event

	combinedEvent chan combinedEvent
	custom        chan *customEvent
}

type combinedEvent struct {
	ev          *socketmode.Event
	customEvent *customEvent
}

func (s *app) Run(handler spanner.EventHandlerFunc) error {
	go func() {
		for ce := range s.custom {
			s.combinedEvent <- combinedEvent{
				customEvent: ce,
			}
		}
	}()
	go func() {
		for evt := range s.events {
			s.combinedEvent <- combinedEvent{
				ev: &evt,
			}
		}
	}()
	go func() {
		for ce := range s.combinedEvent {
			es := parseCombinedEvent(s.client, ce)
			err := handler(es)
			if err != nil {
				log.Printf("handling event: %v", err)
				continue // Move on without acknowledging, will force a repeat
			}
			var req socketmode.Request

			if evt := ce.ev; evt != nil && evt.Request != nil {
				req = *evt.Request
			}

			err = es.finishEvent(request{
				req:    req,
				es:     es,
				hash:   es.hash,
				client: s.client,
			})
			if err != nil {
				log.Printf("handling request: %v", renderSlackError(err))
				continue // Move on without acknowledging, will force a repeat
			}
		}
	}()
	return s.client.RunContext(context.TODO())
}

func (s *app) SendCustom(c spanner.CustomEvent) error {
	s.custom <- &customEvent{
		body: c.Body(),
	}
	return nil
}

type request struct {
	req  socketmode.Request
	es   *event
	hash string

	client socketClient
}

func (r request) Metadata() []byte {
	metadata, err := json.Marshal(r.es.state)
	if err != nil {
		panic(err)
	}
	return metadata
}
