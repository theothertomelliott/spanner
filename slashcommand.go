package chatframework

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
		Title: title,
	}
	return is.ModalInternal
}
