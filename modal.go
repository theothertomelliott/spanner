package chatframework

import (
	"fmt"

	"github.com/slack-go/slack"
)

var _ Modal = &modalSlack{}

type modalSlack struct {
	*BlocksSlack `json:"blocks"` // This ensures that the value is not nil

	Title      string                `json:"title"`
	Submission *modalSubmissionSlack `json:"submission"`
	HasParent  bool                  `json:"has_parent"`

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

func (m *modalSlack) render() *slack.ModalViewRequest {
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

func (m *modalSlack) Submit(text string) ModalSubmission {
	m.submitText = &text
	// This should be redundant, but for some reason, returning m.Submission
	// resulted in `m.Submit("txt") != nil` being false even if m.Submission
	// was nil.
	if m.Submission == nil {
		return nil
	}
	return m.Submission
}

func (m *modalSlack) Close(text string) bool {
	m.closeText = &text
	return m.update == closed
}

func (m *modalSlack) handleRequest(req requestSlack) error {
	var err error

	if m.Submission != nil {
		return m.Submission.handleRequest(req)
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
				return fmt.Errorf("opening view: %w", err)
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
			if slackErr, ok := err.(slack.SlackErrorResponse); ok {
				return fmt.Errorf("updating view: %w %v %v", slackErr, slackErr.ResponseMetadata.Messages, slackErr.ResponseMetadata.Warnings)
			} else {
				return fmt.Errorf("updating view: %T %w", err, err)
			}
		}
	}

	req.client.Ack(req.req, payload)

	return nil
}

func (m *modalSlack) populateEvent(p eventPopulation) error {
	if m.BlocksSlack == nil {
		m.BlocksSlack = &BlocksSlack{}
	}

	if m.Submission != nil {
		return m.Submission.populateEvent(p)
	}

	m.ViewExternalID = p.interactionCallbackEvent.View.ExternalID
	m.ViewID = p.interactionCallbackEvent.View.ID
	m.BlockStates = blockActionToState(p.interactionCallbackEvent.View.State.Values)

	if p.interaction == slack.InteractionTypeBlockActions {
		m.update = action
	}
	if p.interaction == slack.InteractionTypeViewSubmission {
		m.Submission = &modalSubmissionSlack{
			parent: m,
		}
		m.update = submitted
	}
	if p.interaction == slack.InteractionTypeViewClosed {
		m.update = closed
	}

	return nil
}

var _ ModalSubmission = &modalSubmissionSlack{}

type modalSubmissionSlack struct {
	NextModal *modalSlack `json:"next_modal"`

	parent *modalSlack
}

func (m *modalSubmissionSlack) Push(title string) Modal {
	if m.NextModal != nil {
		return m.NextModal
	}

	m.NextModal = &modalSlack{
		BlocksSlack: &BlocksSlack{},
		Title:       title,
		HasParent:   true,
	}
	return m.NextModal
}

func (m *modalSubmissionSlack) Message() Message {
	return nil
}

func (m *modalSubmissionSlack) handleRequest(req requestSlack) error {
	if m.NextModal != nil {
		return m.NextModal.handleRequest(req)
	}

	var payload interface{} = map[string]interface{}{}
	payload = slack.NewClearViewSubmissionResponse()
	req.client.Ack(req.req, payload)

	return nil
}

func (m *modalSubmissionSlack) populateEvent(p eventPopulation) error {
	if m.NextModal != nil {
		return m.NextModal.populateEvent(p)
	}

	return nil
}
