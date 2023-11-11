package spanner

// App is the top level for a chat application.
// Call Run with an event handling function to start the application.
type App interface {
	Run(func(ev Event) error) error
	SendCustom(CustomEvent) error
}

// Event represents an event received from the Slack platform.
// It provides functions representing each type of event that can be received.
// For example, ReceivedMessage will return a message that may have been received in this event.
// Functions will return nil if the current event does not match the type of event.
type Event interface {
	MessageSender

	Connected() bool

	Custom() CustomEvent

	JoinChannel(channel string)

	ReceiveMessage() ReceivedMessage
	SlashCommand(command string) SlashCommand
}

// Metadata provides information common to all events.
type Metadata interface {
	User() User
	Channel() Channel
}

// MessageSender is an interface that can be used to send Slack messages.
type MessageSender interface {
	SendMessage(channelID string) Message
}

// ModalCreator is an interface that can be used to create Slack modal views.
type ModalCreator interface {
	Modal(title string) Modal
}

// SlashCommand represents a received slash command.
// Messages and modal views may be created in response to the command.
type SlashCommand interface {
	Metadata

	MessageSender
	ModalCreator
}

// Modal represents a Slack modal view.
// It can be used to create blocks and handle submission or closing of the modal.
type Modal interface {
	BlockUI
	Submit(title string) ModalSubmission
	Close(title string) bool
}

// ModalSubmission handles a modal being submitted.
// It can be used to send a response message or push a new modal onto the stack.
type ModalSubmission interface {
	MessageSender
	Push(title string) Modal
}

// ReceivedMessage represents a message received from Slack.
type ReceivedMessage interface {
	Metadata
	MessageSender
	Text() string
}

// Message represents a message that can be sent to Slack.
// Messages are constructed using BlockUI commands.
type Message interface {
	BlockUI

	Channel(channelID string)
}
