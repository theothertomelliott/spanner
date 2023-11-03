package chatframework

type slashCommandSlack struct {
	eventMetadataSlack

	*MessageSenderSlack `json:"ms"`

	TriggerID     string      `json:"trigger_id"`
	Command       string      `json:"command"`
	ModalInternal *modalSlack `json:"modal"`
}

var _ SlashCommand = &slashCommandSlack{}

func (is *slashCommandSlack) Modal(title string) Modal {
	if is == nil {
		return nil
	}
	if is.ModalInternal != nil {
		return is.ModalInternal
	}
	is.ModalInternal = &modalSlack{
		BlocksSlack: &BlocksSlack{},
		ChannelID:   is.ChannelInternal,
		Title:       title,
		triggerID:   is.TriggerID,
	}
	return is.ModalInternal
}

func (is *slashCommandSlack) handleRequest(req requestSlack) error {
	err := is.MessageSenderSlack.sendMessages(req)
	if err != nil {
		return err
	}

	if is.ModalInternal != nil {
		return is.ModalInternal.handleRequest(req)
	}
	return nil
}

func (is *slashCommandSlack) populateEvent(p eventPopulation) error {
	err := is.MessageSenderSlack.populateEvent(p)
	if err != nil {
		return err
	}

	if is.ModalInternal != nil {
		return is.ModalInternal.populateEvent(p)
	}
	return nil
}
