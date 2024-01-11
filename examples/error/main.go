package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/theothertomelliott/spanner"
	"github.com/theothertomelliott/spanner/slack"
)

func main() {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	app, err := slack.NewApp(
		slack.AppConfig{
			BotToken:   botToken,
			AppToken:   appToken,
			AckOnError: true,
			FinishInterceptor: func(ctx context.Context, actions []spanner.Action, finish func(context.Context) error) error {
				if len(actions) > 0 {
					var data []interface{}
					for _, action := range actions {
						data = append(data, action.Data())
					}
					dataJson, err := json.MarshalIndent(data, "", "  ")
					if err != nil {
						log.Println("marshalling action data:", err)
					}
					log.Println("Will attempt actions: ", string(dataJson))
				}
				return finish(ctx)
			},
			ActionInterceptor: func(ctx context.Context, action spanner.Action, exec func(context.Context) error) error {
				err := exec(ctx)
				if err != nil {
					dataJson, jsonErr := json.MarshalIndent(action.Data(), "", "  ")
					if jsonErr != nil {
						log.Println("marshalling action data:", err)
					}
					log.Printf("error: %q, when executing action: %v", err, string(dataJson))
				}
				return err
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(func(ctx context.Context, ev spanner.Event) error {
		if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {

			replyGood := ev.SendMessage(msg.Channel().ID())
			replyGood.PlainText("This message should succeed")

			replyBad := ev.SendMessage("invalid_channel")
			replyBad.PlainText("This message will always fail to post")

			replySkipped := ev.SendMessage(msg.Channel().ID())
			replySkipped.PlainText("This message should be skipped because of the previous error")
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
