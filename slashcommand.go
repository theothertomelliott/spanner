package chatframework

import (
	"github.com/slack-go/slack"
)

type slashCommandSlack struct {
	TriggerID     string      `json:"trigger_id"`
	Command       string      `json:"command"`
	ModalInternal *modalSlack `json:"modal"`
}

func (is *slashCommandSlack) Modal(title string) Modal {
	if is == nil {
		return nil
	}
	if is.ModalInternal != nil {
		return is.ModalInternal
	}
	is.ModalInternal = &modalSlack{
		Title:     title,
		triggerID: is.TriggerID,
	}
	return is.ModalInternal
}

func (is *slashCommandSlack) handleRequest(req requestSlack) error {
	if is.ModalInternal != nil {
		return is.ModalInternal.handleRequest(req)
	}
	return nil
}

func (is *slashCommandSlack) populateEvent(interaction slack.InteractionType, view *slack.View) error {
	if is.ModalInternal != nil {
		return is.ModalInternal.populateEvent(interaction, view)
	}
	return nil
}
