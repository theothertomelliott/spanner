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
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = app.SendCustom(context.Background(), slack.NewCustomEvent(map[string]interface{}{
		"field1": "value1",
	}))

	err = app.Run(func(ctx context.Context, ev spanner.Event) {
		if custom := ev.ReceiveCustomEvent(); custom != nil {
			log.Printf("Custom body: %+v", custom.Body())

			msg := ev.SendMessage("C062778EYRZ")
			msg.Markdown(fmt.Sprintf("You sent %+v", custom.Body()))
			input := msg.TextInput("Say something!", "a", "b")
			fmt.Println(input)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
