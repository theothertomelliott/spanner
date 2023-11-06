package slack

import (
	"fmt"

	"github.com/slack-go/slack"
	"github.com/theothertomelliott/chatframework"
)

var _ chatframework.Modal = &modal{}

type modal struct {
	*Blocks `json:"blocks"` // This ensures that the value is not nil

	Title      string           `json:"title"`
	Submission *modalSubmission `json:"submission"`
	HasParent  bool             `json:"has_parent"`
	ChannelID  string           `json:"channel_id"`

	ViewID         string `json:"view_id"`
	ViewExternalID string `json:"view_external_id"`

	triggerID string

	update updateType

	submitText *string
	closeText  *string
}

type updateType int

const (
	created updateType = iota
	action
	submitted
	closed
)

func (m *modal) render() *slack.ModalViewRequest {
	if m == nil {
		return nil
	}
	modal := &slack.ModalViewRequest{
		Type:  slack.ViewType("modal"),
		Title: slack.NewTextBlockObject(slack.PlainTextType, m.Title, false, false),

		Blocks: slack.Blocks{
			BlockSet: m.blocks,
		},
	}

	if m.submitText != nil {
		modal.Submit = slack.NewTextBlockObject(slack.PlainTextType, *m.submitText, false, false)
	}

	if m.closeText != nil {
		modal.Close = slack.NewTextBlockObject(slack.PlainTextType, *m.closeText, false, false)
		modal.NotifyOnClose = true
	}

	return modal
}

func (m *modal) Submit(text string) chatframework.ModalSubmission {
	m.submitText = &text
	// This should be redundant, but for some reason, returning m.Submission
	// resulted in `m.Submit("txt") != nil` being false even if m.Submission
	// was nil.
	if m.Submission == nil {
		return nil
	}
	return m.Submission
}

func (m *modal) Close(text string) bool {
	m.closeText = &text
	return m.update == closed
}

func (m *modal) finishEvent(req request) error {
	var err error

	if m.Submission != nil {
		return m.Submission.finishEvent(req)
	}

	modal := m.render()
	modal.PrivateMetadata = string(req.Metadata())
	fmt.Println("Metadata:", string(modal.PrivateMetadata))
	fmt.Println("Metadata length:", len(modal.PrivateMetadata))

	var payload interface{} = map[string]interface{}{}

	switch update := m.update; update {
	case created:
		if !m.HasParent {
			_, err = req.client.OpenView(m.triggerID, *modal)
			if err != nil {
				return fmt.Errorf("opening view: %w", renderSlackError(err))
			}
		} else {
			payload = slack.NewPushViewSubmissionResponse(modal)
		}
	case action:
		_, err := req.client.UpdateView(
			*modal,
			m.ViewExternalID,
			req.hash,
			m.ViewID,
		)
		if err != nil {
			return fmt.Errorf("updating view: %w", renderSlackError(err))
		}
	}

	req.client.Ack(req.req, payload)

	return nil
}

func (m *modal) populateEvent(p eventPopulation) error {
	if m.Blocks == nil {
		m.Blocks = &Blocks{}
	}

	if m.Submission != nil {
		return m.Submission.populateEvent(p)
	}

	m.ViewExternalID = p.interactionCallbackEvent.View.ExternalID
	m.ViewID = p.interactionCallbackEvent.View.ID
	m.BlockStates = blockActionToState(p)

	if p.interaction == slack.InteractionTypeBlockActions {
		m.update = action
	}
	if p.interaction == slack.InteractionTypeViewSubmission {
		m.Submission = &modalSubmission{
			MessageSender: &MessageSender{
				DefaultChannelID: m.ChannelID,
			},
			parent: m,
		}
		m.update = submitted
	}
	if p.interaction == slack.InteractionTypeViewClosed {
		m.update = closed
	}

	return nil
}

var _ chatframework.ModalSubmission = &modalSubmission{}

type modalSubmission struct {
	*MessageSender `json:"ms"`

	NextModal *modal `json:"next_modal"`

	parent *modal
}

func (m *modalSubmission) Push(title string) chatframework.Modal {
	if m.NextModal != nil {
		return m.NextModal
	}

	m.NextModal = &modal{
		Blocks:    &Blocks{},
		ChannelID: m.parent.ChannelID,
		Title:     title,
		HasParent: true,
	}
	return m.NextModal
}

func (m *modalSubmission) finishEvent(req request) error {
	err := m.MessageSender.sendMessages(req)
	if err != nil {
		return err
	}

	if m.NextModal != nil {
		return m.NextModal.finishEvent(req)
	}

	var payload interface{} = map[string]interface{}{}
	payload = slack.NewClearViewSubmissionResponse()
	req.client.Ack(req.req, payload)

	return nil
}

func (m *modalSubmission) populateEvent(p eventPopulation) error {
	err := m.MessageSender.populateEvent(p)
	if err != nil {
		return err
	}

	if m.NextModal != nil {
		return m.NextModal.populateEvent(p)
	}

	return nil
}
