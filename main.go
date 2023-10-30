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
		fmt.Println("Handling /testslash")
		modal := testSlash.Modal("My Modal")
		modal.Text("Got your slash command")
		tensOptions := []string{}
		for i := 0; i < 10; i++ {
			tensOptions = append(tensOptions, fmt.Sprint(i))
		}
		tensOutput := modal.Select("Tens", tensOptions)
		fmt.Println("Tens:", tensOutput)

		unitsOptions := []string{}
		for i := 0; i < 10; i++ {
			tensPrefix := tensOutput
			if tensPrefix == "0" {
				tensPrefix = ""
			}
			unitsOptions = append(unitsOptions, fmt.Sprintf("%v%v", tensPrefix, i))
		}

		unitsOutput := modal.Select("Units", unitsOptions)
		fmt.Println("Units:", unitsOutput)

	}
	return nil
}
