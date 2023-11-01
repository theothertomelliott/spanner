package chatframework

type slashCommandSlack struct {
	eventMetadataSlack

	TriggerID     string      `json:"trigger_id"`
	Command       string      `json:"command"`
	ModalInternal *modalSlack `json:"modal"`
	Messages      []Message   `json:"messages"`
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
		blocksSlack: &blocksSlack{},
		Title:       title,
		triggerID:   is.TriggerID,
	}
	return is.ModalInternal
}

func (is *slashCommandSlack) Message() Message {
	return nil
}

func (is *slashCommandSlack) handleRequest(req requestSlack) error {
	if is.ModalInternal != nil {
		return is.ModalInternal.handleRequest(req)
	}
	return nil
}

func (is *slashCommandSlack) populateEvent(p eventPopulation) error {
	if is.ModalInternal != nil {
		return is.ModalInternal.populateEvent(p)
	}
	return nil
}
