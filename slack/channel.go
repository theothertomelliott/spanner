package slack

import (
	"context"

	"github.com/slack-go/slack"
	"github.com/theothertomelliott/spanner"
)

var _ spanner.Channel = &channel{}

type channel struct {
	client socketClient

	IDInternal   string `json:"id"`
	NameInternal string `json:"name"`

	Loaded bool `json:"loaded"`
}

func (c *channel) ID() string {
	return c.IDInternal
}

func (c *channel) Name(ctx context.Context) string {
	c.load(ctx)
	return c.NameInternal
}

func (c *channel) load(ctx context.Context) {
	if c.Loaded {
		return
	}

	ch, err := c.client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID: c.IDInternal,
	})
	if err != nil {
		panic(err)
	}
	c.Loaded = true
	c.NameInternal = ch.Name
}
