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

	fmt.Println("You can trigger a modal using the command `/testslash`")
	err = app.Run(handler)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(ev chatframework.Event) error {
	if testSlash := ev.SlashCommand("/testslash"); testSlash != nil {
		fmt.Printf("Handling /testslash from user %v in channel %v\n", testSlash.User(), testSlash.Channel())
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
			if submit := modal.Submit("Submit"); submit != nil {
				fmt.Println("Submitted: ", tensOutput, unitsOutput)

				modal2 := submit.Push("Step 2")
				modal2.Text("Hello")

				dropdown := modal2.Select("Dropdown", []string{"a", "b", "c"})
				fmt.Println("Dropdown:", dropdown)

				singleLine := modal2.TextInput("Single line", "Hint", "Placeholder")
				fmt.Println("Single line:", singleLine)

				multiLine := modal2.MultilineTextInput("Multi line", "Hint", "Placeholder")
				fmt.Println("Multi line:", multiLine)

				if submit := modal2.Submit("Submit"); submit != nil {
					msg := submit.SendMessage()
					msg.Text("Thank you for completing our modal view.")
					msg.Text(fmt.Sprintf("You selected %v, %v, %v", tensOutput, unitsOutput, dropdown))
					msg.Text(fmt.Sprintf("You entered %q, %q", singleLine, multiLine))
				}
			}
		}
	}
	return nil
}
