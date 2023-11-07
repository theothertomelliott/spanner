package slack

import (
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/theothertomelliott/chatframework"
)

var _ chatframework.Channel = &channel{}

type channel struct {
	client *socketmode.Client

	IDInternal   string `json:"id"`
	NameInternal string `json:"name"`

	loaded bool
}

func (c *channel) ID() string {
	return c.IDInternal
}

func (c *channel) Name() string {
	c.load()
	return c.NameInternal
}

func (c *channel) load() {
	if c.loaded {
		return
	}

	ch, err := c.client.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: c.IDInternal,
	})
	if err != nil {
		panic(err)
	}
	c.loaded = true
	c.NameInternal = ch.Name
}
