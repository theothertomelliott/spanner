package spanner

import "context"

type User interface {
	ID() string
	Name(context.Context) string
	RealName(context.Context) string
	Email(context.Context) string
}
