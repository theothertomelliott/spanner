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

		modal.PlainText("Step 1: Choose a number")
		tensOptions := []string{}
		for i := 0; i < 10; i++ {
			tensOptions = append(tensOptions, fmt.Sprint(i))
		}
		tensOutput := modal.Select("Tens", chatframework.Options(tensOptions...))
		fmt.Println("Tens:", tensOutput)

		finalNumber := ""
		if tensOutput != "" {

			unitsOptions := []chatframework.Option{}
			for i := 0; i < 10; i++ {
				tensPrefix := tensOutput
				if tensPrefix == "0" {
					tensPrefix = ""
				}
				unit := chatframework.Option{
					Label:       fmt.Sprint(i),
					Value:       fmt.Sprintf("%v%v", tensPrefix, i),
					Description: fmt.Sprintf("returns %v%v", tensPrefix, i),
				}
				unitsOptions = append(unitsOptions, unit)
			}

			finalNumber = modal.Select("Units", unitsOptions)
			fmt.Println("Final Number:", finalNumber)
		}

		if finalNumber != "" {
			if submit := modal.Submit("Submit"); submit != nil {
				fmt.Println("Your number:", finalNumber)

				modal2 := submit.Push("Step 2")
				modal2.PlainText("Hello")

				dropdown := modal2.Select("Dropdown", chatframework.Options("a", "b", "c"))
				fmt.Println("Dropdown:", dropdown)

				singleLine := modal2.TextInput("Single line", "Hint", "Placeholder")
				fmt.Println("Single line:", singleLine)

				multiLine := modal2.MultilineTextInput("Multi line", "Hint", "Placeholder")
				fmt.Println("Multi line:", multiLine)

				if submit := modal2.Submit("Submit"); submit != nil {
					msg := submit.SendMessage()
					msg.PlainText("Thank you for completing our modal view.")
					msg.PlainText(fmt.Sprintf("Your number was %v", finalNumber))
					msg.PlainText(fmt.Sprintf("You entered %v, %q and %q in the second view", dropdown, singleLine, multiLine))
				}
			}
		}
	}
	return nil
}
