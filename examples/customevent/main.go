package main

import (
	"fmt"
	"log"
	"os"

	"github.com/theothertomelliott/spanner"
	"github.com/theothertomelliott/spanner/slack"
)

func main() {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	app, err := slack.NewApp(botToken, appToken)
	if err != nil {
		log.Fatal(err)
	}

	_ = app.SendCustom(slack.NewCustomEvent(map[string]interface{}{
		"field1": "value1",
	}))

	err = app.Run(func(ev spanner.Event) error {
		if custom := ev.Custom(); custom != nil {
			log.Printf("Custom body: %+v", custom.Body())
		}
		if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {

			reply := msg.SendMessage(msg.Channel().ID())
			reply.PlainText(fmt.Sprintf("Hello to you too: %v", msg.User()))

			letter := reply.Select("Pick a letter", spanner.Options("a", "b", "c"))
			if letter != "" {
				msg.SendMessage(msg.Channel().ID()).PlainText(fmt.Sprintf("You chose %q", letter))
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
