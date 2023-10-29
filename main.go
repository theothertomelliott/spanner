package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	app, err := NewSlackApp(botToken, appToken)
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(handler)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(ev EventState) error {
	if msg := ev.ReceiveMessage(); msg != nil {
		fmt.Println("Received a message:", msg.Text)
		if msg.Text == "hello" {
			// TODO: Send a reply message
			fmt.Println("got a hello")
		}
	}
	if testSlash := ev.SlashCommand("/testslash"); testSlash != nil {
		fmt.Println("Slash command received")
		modal := testSlash.Modal("My Modal")
		modal.Text("Got your slash command")
	}
	return nil
}
