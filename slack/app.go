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

	return newAppWithClient(&wrappedClient{
		Client: client,
	}, events), nil
}

func newAppWithClient(client socketClient, slackEvents chan socketmode.Event) spanner.App {
	return &app{
		client:        client,
		slackEvents:   slackEvents,
		combinedEvent: make(chan combinedEvent, 2),
		customEvents:  make(chan *customEvent, 2),
	}
}

type wrappedClient struct {
	*socketmode.Client
}

func (w *wrappedClient) SendMessageWithMetadata(ctx context.Context, channelID string, blocks []slack.Block, metadata slack.SlackMetadata) (string, string, string, error) {
	return w.SendMessageContext(ctx, channelID, slack.MsgOptionBlocks(blocks...), slack.MsgOptionMetadata(metadata))
}

func (w *wrappedClient) UpdateMessageWithMetadata(ctx context.Context, channelID string, timestamp string, blocks []slack.Block, metadata slack.SlackMetadata) (string, string, string, error) {
	return w.UpdateMessageContext(ctx, channelID, timestamp, slack.MsgOptionBlocks(blocks...), slack.MsgOptionMetadata(metadata))
}

type app struct {
	client socketClient

	slackEvents   chan socketmode.Event
	customEvents  chan *customEvent
	combinedEvent chan combinedEvent
}

type combinedEvent struct {
	ev          *socketmode.Event
	customEvent *customEvent
}

func (s *app) Run(handler spanner.EventHandlerFunc) error {
	go func() {
		for ce := range s.customEvents {
			s.combinedEvent <- combinedEvent{
				customEvent: ce,
			}
		}
	}()
	go func() {
		for evt := range s.slackEvents {
			s.combinedEvent <- combinedEvent{
				ev: &evt,
			}
		}
	}()

	done := make(chan error)
	go func() {
		err := s.client.RunContext(context.TODO())
		if err != nil {
			done <- err
		}
		close(done)
	}()

	for {
		select {
		case ce := <-s.combinedEvent:
			s.handleEvent(handler, ce)
		case err := <-done:
			return err
		}
	}
}

func (s *app) handleEvent(handler spanner.EventHandlerFunc, ce combinedEvent) {
	es := parseCombinedEvent(s.client, ce)
	err := handler(es)
	if err != nil {
		return // Move on without acknowledging, will force a repeat
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
		return // Move on without acknowledging, will force a repeat
	}
}

func (s *app) SendCustom(c spanner.CustomEvent) error {
	s.customEvents <- &customEvent{
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
