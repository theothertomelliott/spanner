package main

import (
	"fmt"
	"log"
	"os"

	"github.com/theothertomelliott/chatframework"
)

func main() {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	app, err := chatframework.NewSlackApp(botToken, appToken)
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(handler)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(ev chatframework.EventState) error {
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

		if modal.Close("Cancel") {
			fmt.Println("Closed")
			return nil
		}

		modal.Text("Got your slash command")
		tensOptions := []string{}
		for i := 0; i < 10; i++ {
			tensOptions = append(tensOptions, fmt.Sprint(i))
		}
		tensOutput := modal.Select("Tens", tensOptions)
		fmt.Println("Tens:", tensOutput)

		unitsOutput := ""
		if tensOutput != "" {

			unitsOptions := []string{}
			for i := 0; i < 10; i++ {
				tensPrefix := tensOutput
				if tensPrefix == "0" {
					tensPrefix = ""
				}
				unitsOptions = append(unitsOptions, fmt.Sprintf("%v%v", tensPrefix, i))
			}

			unitsOutput = modal.Select("Units", unitsOptions)
			fmt.Println("Units:", unitsOutput)
		}

		if unitsOutput != "" {
			if modal.Submit("Submit") {
				fmt.Println("Submitted: ", tensOutput, unitsOutput)

				modal2 := modal.Push("Step 2")
				modal2.Text("Hello")

				dropdown := modal2.Select("Dropdown", []string{"a", "b", "c"})
				fmt.Println("Dropdown:", dropdown)

				if modal2.Submit("Submit") {
					fmt.Println("Final submission")
				}
			}
		}
	}
	return nil
}
