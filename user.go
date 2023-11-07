package chatframework

type User interface {
	ID() string
	Name() string
	RealName() string
	Email() string
}
