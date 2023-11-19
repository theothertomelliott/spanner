package main

import (
	"fmt"
	"log"
	"os"

	spanner "github.com/theothertomelliott/spanner"
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

	fmt.Println("Starting app")
	err = app.Run(func(ev spanner.Event) error {
		if ev.ReceiveConnected() {
			log.Println("Connected - will do some setup here")
			ev.JoinChannel("C062778EYRZ") // Test channel from Slack workspace used for QA
			ev.JoinChannel("C016E85HUUA") // #random channel from QA workspace
		}

		if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {

			reply := ev.SendMessage(msg.Channel().ID())
			reply.Markdown(fmt.Sprintf("Hello, *%v*", msg.User().RealName()))

			reply.PlainText("Here are examples of supported block UI elements")

			reply.Markdown("This is a *markdown* message")

			reply.Divider()

			reply.Header("Text inputs")

			singleLine := reply.TextInput("Single line", "Enter a single line", "Placeholder")
			multiLine := reply.MultilineTextInput("Multi line", "Enter a multi line", "Placeholder")

			reply.Divider()

			reply.Header("Select inputs")

			letter := reply.Select("Pick a letter", spanner.Options("a", "b", "c"))
			numbers := reply.MultipleSelect("Pick some numbers", spanner.Options("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"))

			if reply.Button("Done") {
				summary := ev.SendMessage(msg.Channel().ID())
				summary.PlainText("Here's a summary of what you entered")
				summary.PlainText(fmt.Sprintf("Original poster: %v", msg.User().RealName()))
				summary.PlainText(fmt.Sprintf("Single line: %q", singleLine))
				summary.PlainText(fmt.Sprintf("Multi line: %q", multiLine))
				summary.PlainText(fmt.Sprintf("You chose %q", letter))
				summary.PlainText(fmt.Sprintf("Numbers: %v", numbers))
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
