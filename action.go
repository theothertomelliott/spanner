package spanner

type Action interface {
	Type() string
	Data() interface{}
}
