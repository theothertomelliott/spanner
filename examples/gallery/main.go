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
			reply.PlainText("Here are examples of supported block UI elements")

			reply.Markdown("This is a *markdown* message")

			reply.Divider()

			reply.Header("Text inputs")

			singleLine := reply.TextInput("Single line", "Enter a single line", "Placeholder")
			multiLine := reply.MultilineTextInput("Multi line", "Enter a multi line", "Placeholder")

			reply.Divider()

			reply.Header("Select inputs")

			letter := reply.Select("Pick a letter", chatframework.Options("a", "b", "c"))
			numbers := reply.MultipleSelect("Pick some numbers", chatframework.Options("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"))

			if reply.Button("Done") {
				summary := msg.SendMessage()
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
