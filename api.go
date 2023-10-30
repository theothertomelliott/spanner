package chatframework

type App interface {
	Run(func(ev EventState) error) error
}

type EventState interface {
	ReceiveMessage() *Message
	SlashCommand(command string) SlashCommand
}

type Interaction interface {
	// TODO: Refactor so that interactions create a separate object for handling modals, messages, etc
	Modal(string) Modal
}

type SlashCommand interface {
	Interaction
}

type Modal interface {
	Text(message string)
	Select(title string, options []string) string
	Submit(title string) bool
	Close(title string) bool
}

type Message struct {
	UserID string
	Text   string
}
