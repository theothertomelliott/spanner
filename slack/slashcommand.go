package slack

import (
	"context"

	"github.com/theothertomelliott/spanner"
)

type slashCommand struct {
	actionQueue *actionQueue

	eventMetadata
	ephemeralSender

	TriggerID     string `json:"trigger_id"`
	Command       string `json:"command"`
	ModalInternal *modal `json:"modal"`
}

var _ spanner.SlashCommand = &slashCommand{}
var _ eventPopulator = &slashCommand{}

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
	is.actionQueue.enqueue(is.ModalInternal)
	return is.ModalInternal
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
