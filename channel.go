package spanner

import "context"

type Channel interface {
	ID() string
	Name(context.Context) string
}
