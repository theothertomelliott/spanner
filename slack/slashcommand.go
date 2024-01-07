package slack

import (
	"context"

	"github.com/theothertomelliott/spanner"
)

type slashCommand struct {
	eventMetadata
	ephemeralSender

	TriggerID     string `json:"trigger_id"`
	Command       string `json:"command"`
	ModalInternal *modal `json:"modal"`
}

var _ spanner.SlashCommand = &slashCommand{}
var _ eventPopulator = &slashCommand{}
var _ eventFinisher = &slashCommand{}

func (is *slashCommand) Modal(title string) spanner.Modal {
	if is == nil {
		return nil
	}
	if is.ModalInternal != nil {
		return is.ModalInternal
	}
	is.ModalInternal = &modal{
		Blocks:    &Blocks{},
		ChannelID: is.ChannelInfo.IDInternal,
		Title:     title,
		triggerID: is.TriggerID,
	}
	return is.ModalInternal
}

func (is *slashCommand) finishEvent(ctx context.Context, req request) error {
	if is.ModalInternal != nil {
		return is.ModalInternal.finishEvent(ctx, req)
	}
	if is.ephemeralSender.Text != nil {
		return is.ephemeralSender.finishEvent(ctx, req)
	}

	var payload interface{} = map[string]interface{}{}
	req.client.Ack(req.req, payload)

	return nil
}

func (is *slashCommand) populateEvent(ctx context.Context, p eventPopulation, depth int) error {
	if is.ModalInternal != nil {
		return is.ModalInternal.populateEvent(ctx, p, depth+1)
	}
	if is.ephemeralSender.Text != nil {
		return is.ephemeralSender.populateEvent(ctx, p, depth+1)
	}
	return nil
}
