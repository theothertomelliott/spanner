package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/theothertomelliott/spanner"
)

var _ spanner.Modal = &modal{}
var _ eventPopulator = &modal{}
var _ action = &modal{}

type modal struct {
	actionQueue *actionQueue

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
	modalUpdateCreated updateType = iota
	modalUpdateAction
	modalUpdateSubmitted
	modalUpdateClosed
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

func (m *modal) SubmitButton(text string) spanner.ModalSubmission {
	m.submitText = &text
	// This should be redundant, but for some reason, returning m.Submission
	// resulted in `m.Submit("txt") != nil` being false even if m.Submission
	// was nil.
	if m.Submission == nil {
		return nil
	}
	return m.Submission
}

func (m *modal) CloseButton(text string) bool {
	m.closeText = &text
	return m.update == modalUpdateClosed
}

func (m *modal) exec(ctx context.Context, req request) (interface{}, error) {
	var err error

	modal := m.render()
	modal.PrivateMetadata = string(req.Metadata())

	var payload interface{} = map[string]interface{}{}

	switch update := m.update; update {
	case modalUpdateCreated:
		if !m.HasParent {
			_, err = req.client.OpenViewContext(ctx, m.triggerID, *modal)
			if err != nil {
				return nil, fmt.Errorf("opening view: %w", renderSlackError(err))
			}
		} else {
			payload = slack.NewPushViewSubmissionResponse(modal)
		}
	case modalUpdateAction:
		_, err := req.client.UpdateViewContext(
			ctx,
			*modal,
			m.ViewExternalID,
			req.hash,
			m.ViewID,
		)
		if err != nil {
			return nil, fmt.Errorf("updating view: %w", renderSlackError(err))
		}
	}

	return payload, nil
}

func (m *modal) populateEvent(ctx context.Context, p eventPopulation, depth int) error {
	if m.Blocks == nil {
		m.Blocks = &Blocks{}
	}

	if m.Submission != nil {
		return m.Submission.populateEvent(ctx, p, depth+1)
	}

	m.actionQueue = p.actionQueue

	m.ViewExternalID = p.interactionCallbackEvent.View.ExternalID
	m.ViewID = p.interactionCallbackEvent.View.ID
	m.BlockStates = blockActionToState(p)

	if p.interaction == slack.InteractionTypeBlockActions {
		m.update = modalUpdateAction
		m.actionQueue.enqueue(m)
	}
	if p.interaction == slack.InteractionTypeViewSubmission {
		m.Submission = &modalSubmission{
			actionQueue: p.actionQueue,
			parent:      m,
		}
		m.update = modalUpdateSubmitted
		m.actionQueue.enqueue(m.Submission)
	}
	if p.interaction == slack.InteractionTypeViewClosed {
		m.update = modalUpdateClosed
	}

	return nil
}

func (*modal) Type() string {
	return "modal"
}

func (m *modal) Data() interface{} {
	// TODO: This should be more well-defined
	return map[string]interface{}{
		"title":            m.Title,
		"blocks":           m.blocks,
		"channel_id":       m.ChannelID,
		"view_id":          m.ViewID,
		"view_id_external": m.ViewExternalID,
	}
}

var _ spanner.ModalSubmission = &modalSubmission{}
var _ eventPopulator = &modalSubmission{}

type modalSubmission struct {
	actionQueue *actionQueue

	NextModal *modal `json:"next_modal"`

	parent *modal
}

func (m *modalSubmission) PushModal(title string) spanner.Modal {
	if m.NextModal != nil {
		return m.NextModal
	}

	m.NextModal = &modal{
		Blocks:    &Blocks{},
		ChannelID: m.parent.ChannelID,
		Title:     title,
		HasParent: true,
	}
	m.actionQueue.enqueue(m.NextModal)
	return m.NextModal
}

func (m *modalSubmission) exec(ctx context.Context, req request) (interface{}, error) {
	var payload interface{} = map[string]interface{}{}
	payload = slack.NewClearViewSubmissionResponse()
	return payload, nil
}

func (m *modalSubmission) populateEvent(ctx context.Context, p eventPopulation, depth int) error {
	if m.NextModal != nil {
		return m.NextModal.populateEvent(ctx, p, depth+1)
	}

	return nil
}

func (*modalSubmission) Type() string {
	return "modal-submission"
}

func (ms *modalSubmission) Data() interface{} {
	// TODO: This should be more well defined
	return map[string]interface{}{
		"next_modal": ms.NextModal,
	}
}
