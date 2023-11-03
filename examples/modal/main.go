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

func handler(ev chatframework.Event) error {
	if msg := ev.ReceiveMessage(); msg != nil {
		fmt.Println("Received a message:", msg.Text())
		if msg.Text() == "hello" {
			outMessage := msg.SendMessage()
			outMessage.Text(fmt.Sprintf("Hello to you too: %v", msg.User()))
			selectValue := outMessage.Select("Select", []string{"a", "b", "c"})

			outMessage.Divider()

			select2Value := outMessage.Select("Select 2", []string{"d", "e", "f"})

			out2 := msg.SendMessage()
			out2.Text("Here's another message for good measure")

			if select2Value != "" && selectValue != "" {
				result := msg.SendMessage()
				result.Text(fmt.Sprintf("You selected %v and %v", selectValue, select2Value))
			}
		}
	}
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
					msg.Text(fmt.Sprintf("You selected %v, %v, %v", tensOutput, unitsOutput, dropdown))
					msg.Text(fmt.Sprintf("You entered %q, %q", singleLine, multiLine))
				}
			}
		}
	}
	return nil
}
