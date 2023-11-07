package chatframework

type Channel interface {
	ID() string
	Name() string
}
