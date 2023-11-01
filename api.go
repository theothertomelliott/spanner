package chatframework

type App interface {
	Run(func(ev Event) error) error
}

type Event interface {
	ReceiveMessage() ReceivedMessage
	SlashCommand(command string) SlashCommand
}

type Interaction interface {
	Modal(title string) Modal
	Message() Message
}

type SlashCommand interface {
	Metadata
	Interaction
}

type BlockUI interface {
	Text(message string)
	Select(title string, options []string) string
}

type Modal interface {
	BlockUI
	Submit(title string) ModalSubmission
	Close(title string) bool
}

type ModalSubmission interface {
	Push(title string) Modal
	Message() Message
}

type Metadata interface {
	User() string
	Channel() string
}

type ReceivedMessage interface {
	Metadata
	Text() string
	SendMessage() Message
}

type Message interface {
	BlockUI

	Channel(channelID string)
}
