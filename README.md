# Spanner for Slack

Spanner is a Go framework for building interactive Slack applications.

The framework adopts a model inspired by [Streamlit](https://streamlit.io/) with a single event handling loop to construct your UI and to respond to user input - so you can iterate faster and focus on your business logic!

## Getting Started

### Create a Slack app

In order to develop an app, you'll need to create one in your Slack workspace. You can find instructions
for creating a Socket mode app at: https://api.slack.com/apis/connections/socket#creating

### Setup

Import the framework into your project with `go get``:

    go get github.com/theothertomelliott/spanner

Then import the API and Slack application package into your app:

```
import (
  "github.com/theothertomelliott/spanner"
  "github.com/theothertomelliott/spanner/slack"
)
```

### Your application

Now you can define an application and run it:

```
app, err := slack.NewApp(
    slack.AppConfig{
        BotToken: botToken,
        AppToken: appToken,
    },
)
if err != nil {
    log.Fatal(err)
}

err = app.Run(func(ctx context.Context, ev spanner.Event) {
    // TODO: Handle events here
    return nil
})
if err != nil {
    log.Fatal(err)
}
```

Where `botToken` and `appToken` are the tokens you created for your Slack application.

### Handling Events

The function passed in to `app.Run` is your event handling function, and will be called every time Slack
sends an event to your app. In the example above, our app will ignore every event, so let's do something
with messages coming in.

```
err = app.Run(func(ctx context.Context, ev spanner.Event) {
    if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {
        reply := msg.SendMessage()
        reply.Text(fmt.Sprintf("Hello to you too: %v", msg.User()))
    }
})
```

The code above will listen for messages in any channel your bot is in, and if the text of a messge is
exactly "hello", will respond with a greeting.

### UI Elements

You can also easily add UI elements to your messages. Let's add a dropdown to our message and do something
with the option the user chooses.

```
err = app.Run(func(ctx context.Context, ev spanner.Event) {
    if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {

        reply := ev.SendMessage()
        reply.Text(fmt.Sprintf("Hello to you too: %v", msg.User()))

        letter := reply.Select("Pick a letter", spanner.SelectOptions("a", "b", "c"))
        if letter != "" {
            ev.SendMessage().Text(fmt.Sprintf("You chose %q", letter))
        }
    }
})
```

## Event Lifecycle

Events received by a Spanner app go through 2 phases: Handling and Finishing.

*Handling* is when your handler function is called, which uses the event to specify how to
respond. Calls to perform actions like `SendMessage` or `JoinChannel` are deferred until 
the Finishing phase.

*Finishing* is when actions are actually performed in the order they were declared in the Handling phase.

## Custom Events

You can send custom events to your Spanner event handler to allow for use cases like cron tasks or sending
message in response to third-party events.

The `SendCustom` function allows you to send an event with an arbitrary `map[string]interface{}` payload:

```
_ = app.SendCustom(context.Background(), slack.NewCustomEvent(map[string]interface{}{
    "field1": "value1",
}))
```

This event may then be received in your handler, and you can send messages in response:

```
if custom := ev.ReceiveCustomEvent(); custom != nil {
    msg := ev.SendMessage("C062778EYRZ")
    msg.Markdown(fmt.Sprintf("You sent %+v", custom.Body()))
}
```

## Error Handling

The handler function may not return an error. If you call functions that may error out during handling, it
is recommended to provide feedback to your user via messages and other interactive elements.

Because actions are not performed until after your handler function returns, error handling for these actions
can be deferred by specifying a callback function. Messages and other action-related types have an `ErrorFunc` 
function that allows you to specify this callback:

```
badMessage := ev.SendMessage("invalid_channel")
badMessage.PlainText("This message will always fail to post")
badMessage.ErrorFunc(func(ctx context.Context, ev spanner.ErrorEvent) {
    errorNotice := ev.SendMessage(msg.Channel().ID())
    errorNotice.PlainText(fmt.Sprintf("There was an error sending a message: %v", ev.ReceiveError()))
})
```

This function is called during the Finishing phase when an action fails, this effectively starts a new event
cycle so you can send messages to report the error. When an action fails, all subsequent actions for the current
event are aborted.

## Interceptors

You can specify interceptors to capture lifecycle events, which allows you to add common logging, tracing or other
instrumentation to your event handling.

The interceptors are specified as parameters of your app config:

```
slack.AppConfig{
    BotToken:   botToken,
    AppToken:   appToken,
    // ...
    EventInterceptor: func(ctx context.Context, process func(context.Context)) {
        log.Println("Event received")
        process(ctx)
    },
    HandlerInterceptor: func(ctx context.Context, eventType string, handle func(context.Context)) {
        log.Println("Handling event type: ", eventType)
        handle(ctx)
    },
    FinishInterceptor: func(ctx context.Context, actions []spanner.Action, finish func(context.Context) error) error {
        log.Printf("Finishing with %d actions", len(actions))
        return finish(ctx)
    },
    ActionInterceptor: func(ctx context.Context, action spanner.Action, exec func(context.Context) error) error {
        log.Println("Performing action: ", action.Type())
        return exec(ctx)
    },
},
```

Interceptors cannot influence the specific handling done or actions performed, but it can abort a step in event handling
by not calling the provided function.

## Examples

A set of examples can be found in the [examples directory](./examples).
