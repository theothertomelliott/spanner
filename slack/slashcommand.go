package slack

import "github.com/theothertomelliott/chatframework"

type slashCommand struct {
	eventMetadata

	*MessageSender `json:"ms"`

	TriggerID     string `json:"trigger_id"`
	Command       string `json:"command"`
	ModalInternal *modal `json:"modal"`
}

var _ chatframework.SlashCommand = &slashCommand{}

func (is *slashCommand) Modal(title string) chatframework.Modal {
	if is == nil {
		return nil
	}
	if is.ModalInternal != nil {
		return is.ModalInternal
	}
	is.ModalInternal = &modal{
		Blocks:    &Blocks{},
		ChannelID: is.ChannelInternal,
		Title:     title,
		triggerID: is.TriggerID,
	}
	return is.ModalInternal
}

func (is *slashCommand) handleRequest(req request) error {
	err := is.MessageSender.sendMessages(req)
	if err != nil {
		return err
	}

	if is.ModalInternal != nil {
		return is.ModalInternal.handleRequest(req)
	}
	return nil
}

func (is *slashCommand) populateEvent(p eventPopulation) error {
	err := is.MessageSender.populateEvent(p)
	if err != nil {
		return err
	}

	if is.ModalInternal != nil {
		return is.ModalInternal.populateEvent(p)
	}
	return nil
}