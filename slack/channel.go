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

var _ action = &joinChannelAction{}

type joinChannelAction struct {
	channelID string
	errFunc   spanner.ErrorFunc
}

func (j *joinChannelAction) ErrorFunc(ef spanner.ErrorFunc) {
	j.errFunc = ef
}

func (j *joinChannelAction) getErrorFunc() spanner.ErrorFunc {
	return j.errFunc
}

// Data implements action.
func (j *joinChannelAction) Data() interface{} {
	// TODO: This should be more well-defined
	return map[string]interface{}{
		"channel_id": j.channelID,
	}
}

// Type implements action.
func (*joinChannelAction) Type() string {
	return "join_channel"
}

// exec implements action.
func (a *joinChannelAction) exec(ctx context.Context, req request) (interface{}, error) {
	_, _, _, err := req.client.JoinConversationContext(ctx, a.channelID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
