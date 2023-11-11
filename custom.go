package spanner

type CustomEvent interface {
	Body() map[string]interface{}
}
