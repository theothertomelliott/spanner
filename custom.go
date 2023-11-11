package spanner

type CustomEvent interface {
	MessageSender

	Body() map[string]interface{}
}
