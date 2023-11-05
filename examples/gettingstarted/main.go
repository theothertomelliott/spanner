package main

import (
	"fmt"
	"log"
	"os"

	"github.com/theothertomelliott/chatframework"
	"github.com/theothertomelliott/chatframework/slack"
)

func main() {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	app, err := slack.NewApp(botToken, appToken)
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(func(ev chatframework.Event) error {
		if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {

			reply := msg.SendMessage()
			reply.Text(fmt.Sprintf("Hello to you too: %v", msg.User()))

			letter := reply.Select("Pick a letter", []string{"a", "b", "c"})
			if letter != "" {
				msg.SendMessage().Text(fmt.Sprintf("You chose %q", letter))
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
