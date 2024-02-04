package main

import (
	"context"
	"fmt"
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
			BotToken: botToken,
			AppToken: appToken,
			Debug:    true,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(func(ctx context.Context, ev spanner.Event) {
		if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {

			reply := ev.SendMessage(msg.Channel().ID())
			reply.PlainText(fmt.Sprintf("Hello to you too: %v", msg.User()))

			letter := reply.Select("Pick a letter", spanner.Options("a", "b", "c"))
			if letter != "" {
				ev.SendMessage(msg.Channel().ID()).PlainText(fmt.Sprintf("You chose %q", letter))
			}
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
