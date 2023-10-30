package chatframework

type App interface {
	Run(func(ev EventState) error) error
}

type EventState interface {
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
	Submit(title string) bool
	Close(title string) bool

	Push(title string) Modal
}

type Message struct {
	UserID string
	Text   string
}
