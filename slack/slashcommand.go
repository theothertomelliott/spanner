package slack

import "github.com/theothertomelliott/spanner"

type slashCommand struct {
	eventMetadata

	*MessageSender `json:"ms"`

	TriggerID     string `json:"trigger_id"`
	Command       string `json:"command"`
	ModalInternal *modal `json:"modal"`
}

var _ spanner.SlashCommand = &slashCommand{}

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

func (is *slashCommand) finishEvent(req request) error {
	err := is.MessageSender.sendMessages(req)
	if err != nil {
		return err
	}

	if is.ModalInternal != nil {
		return is.ModalInternal.finishEvent(req)
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
