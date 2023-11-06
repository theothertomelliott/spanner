package slack

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/chatframework"
)

// NewApp creates a new slack app.
//
// botToken is the token for the bot user, with prefix 'xoxb-'
// appToken is the token for the app, with prefix 'xapp-'
//
// Slack apps use socket mode to handle events, so the app will need to be configured to use socket mode.
// https://api.slack.com/apis/connections/socket
//
// As at November 2023, this means that these apps cannot be distributed in the public Slack app directory.
func NewApp(botToken string, appToken string) (*app, error) {
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

	return &app{
		api:    api,
		client: client,
	}, nil
}

type app struct {
	api    *slack.Client
	client *socketmode.Client
}

func (s *app) Run(handler func(ev chatframework.Event) error) error {
	go func() {
		for evt := range s.client.Events {
			es := parseSlackEvent(evt)
			err := handler(es)
			if err != nil {
				log.Printf("handling event: %v", err)
				continue // Move on without acknowledging, will force a repeat
			}
			var req socketmode.Request

			if evt.Request != nil {
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
	return s.client.Run()
}

type request struct {
	req  socketmode.Request
	es   *event
	hash string

	client *socketmode.Client
}

func (r request) Metadata() []byte {
	metadata, err := json.Marshal(r.es.state)
	if err != nil {
		panic(err)
	}
	return metadata
}
