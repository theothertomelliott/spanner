package spanner

import "context"

// App is the top level for a chat application.
// Call Run with an event handling function to start the application.
type App interface {
	Run(EventHandlerFunc) error
	SendCustom(context.Context, CustomEvent) error
	SetPostEventFunc(PostEventFunc)
}

// EventHandlerFunc represents a function that processes chat events from Spanner.
// This function will be called multiple times and is responsible both for creating
// UI elements and responding to the input received.
type EventHandlerFunc func(context.Context, Event) error

// PostEventFunc represents a function that is called after an event is procesed by a
// Spanner app.
type PostEventFunc func(context.Context)

// Event represents an event received from the Slack platform.
// It provides functions representing each type of event that can be received.
// For example, ReceivedMessage will return a message that may have been received in this event.
// Functions will return nil if the current event does not match the type of event.
type Event interface {
	ReceiveConnected() bool
	ReceiveCustomEvent() CustomEvent
	ReceiveMessage() ReceivedMessage
	ReceiveSlashCommand(command string) SlashCommand

	JoinChannel(channelID string)
	SendMessage(channelID string) Message
}

// Metadata provides information common to all events.
type Metadata interface {
	User() User
	Channel() Channel
}

// ModalCreator is an interface that can be used to create Slack modal views.
type ModalCreator interface {
	Modal(title string) Modal
}

// SlashCommand represents a received slash command.
// Messages and modal views may be created in response to the command.
type SlashCommand interface {
	EphemeralSender
	Metadata
	ModalCreator
}

// Modal represents a Slack modal view.
// It can be used to create blocks and handle submission or closing of the modal.
type Modal interface {
	BlockUI
	SubmitButton(title string) ModalSubmission
	CloseButton(title string) bool
}

// ModalSubmission handles a modal being submitted.
// It can be used to send a response message or push a new modal onto the stack.
type ModalSubmission interface {
	PushModal(title string) Modal
}

// ReceivedMessage represents a message received from Slack.
type ReceivedMessage interface {
	Metadata
	Text() string
}

type EphemeralSender interface {
	SendEphemeralMessage(text string)
}

// Message represents a message that can be sent to Slack.
// Messages are constructed using BlockUI commands.
type Message interface {
	BlockUI

	Channel(channelID string)
}
