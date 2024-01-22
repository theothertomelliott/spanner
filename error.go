package spanner

import "context"

type HasError interface {
	ErrorFunc(ErrorFunc)
}

type ErrorFunc func(ctx context.Context, ev ErrorEvent)

type ErrorEvent interface {
	SendMessage(channelID string) ErrorMessage
	ReceiveError() error
}

type ErrorMessage interface {
	NonInteractiveMessage
}
