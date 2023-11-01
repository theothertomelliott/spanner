package chatframework

type App interface {
	Run(func(ev Event) error) error
}

type Event interface {
	ReceiveMessage() *Message
	SlashCommand(command string) SlashCommand
}

type Interaction interface {
	Modal(title string) Modal
}

type SlashCommand interface {
	Interaction
}

type Modal interface {
	Text(message string)
	Select(title string, options []string) string
	Submit(title string) ModalSubmission
	Close(title string) bool
}

type ModalSubmission interface {
	Push(title string) Modal
}

type Message struct {
	UserID string
	Text   string
}
