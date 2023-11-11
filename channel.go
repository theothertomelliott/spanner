package spanner

type Channel interface {
	ID() string
	Name() string
}
