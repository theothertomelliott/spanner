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
app, err := slack.NewApp(botToken, appToken)
if err != nil {
    log.Fatal(err)
}

err = app.Run(func(ev Event) error {
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
err = app.Run(func(ev spanner.Event) error {
    if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {
        reply := msg.SendMessage()
        reply.Text(fmt.Sprintf("Hello to you too: %v", msg.User()))
    }
    return nil
})
```

The code above will listen for messages in any channel your bot is in, and if the text of a messge is
exactly "hello", will respond with a greeting.

### UI Elements

You can also easily add UI elements to your messages. Let's add a dropdown to our message and do something
with the option the user chooses.

```
err = app.Run(func(ev spanner.Event) error {
    if msg := ev.ReceiveMessage(); msg != nil && msg.Text() == "hello" {

        reply := ev.SendMessage()
        reply.Text(fmt.Sprintf("Hello to you too: %v", msg.User()))

		letter := reply.Select("Pick a letter", spanner.SelectOptions("a", "b", "c"))
        if letter != "" {
            ev.SendMessage().Text(fmt.Sprintf("You chose %q", letter))
        }
    }
    return nil
})
```

## Examples

A set of examples can be found in the [examples directory](./examples).
