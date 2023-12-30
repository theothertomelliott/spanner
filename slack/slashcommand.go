package slack

import "github.com/theothertomelliott/spanner"

type slashCommand struct {
	eventMetadata
	ephemeralSender

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
	if is.ModalInternal != nil {
		return is.ModalInternal.finishEvent(req)
	}
	if is.ephemeralSender.Text != nil {
		return is.ephemeralSender.finishEvent(req)
	}

	var payload interface{} = map[string]interface{}{}
	req.client.Ack(req.req, payload)

	return nil
}

func (is *slashCommand) populateEvent(p eventPopulation) error {
	if is.ModalInternal != nil {
		return is.ModalInternal.populateEvent(p)
	}
	if is.ephemeralSender.Text != nil {
		return is.ephemeralSender.populateEvent(p)
	}
	return nil
}
