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

	app, err := slack.NewApp(
		slack.AppConfig{
			BotToken: botToken,
			AppToken: appToken,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = app.SendCustom(slack.NewCustomEvent(map[string]interface{}{
		"field1": "value1",
	}))

	err = app.Run(func(ev spanner.Event) error {
		if custom := ev.Custom(); custom != nil {
			log.Printf("Custom body: %+v", custom.Body())

			msg := ev.SendMessage("C062778EYRZ")
			msg.Markdown(fmt.Sprintf("You sent %+v", custom.Body()))
			input := msg.TextInput("Say something!", "a", "b")
			fmt.Println(input)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
