package spanner

import "context"

type Action interface {
	HasError

	Type() string
	Data() interface{}
}

type EventInterceptor func(ctx context.Context, process func(context.Context))

type HandlerInterceptor func(ctx context.Context, eventType string, handle func(context.Context))

// ActionInterceptor intercepts a single action to allow for instrumentation.
// The next function must be called to perform the action itself.
// The configuration for the action cannot be changed.
type ActionInterceptor func(ctx context.Context, action Action, next func(context.Context) error) error

// FinishInterceptor intercepts the finishing step of handling an event to allow for instrumentation.
// The finish function must be called to perform the required actions.
// The set of actions cannot be changed.
type FinishInterceptor func(ctx context.Context, actions []Action, finish func(context.Context) error) error
