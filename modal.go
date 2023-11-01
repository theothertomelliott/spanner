package chatframework

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type modalSlack struct {
	Title      string                `json:"title"`
	Submission *modalSubmissionSlack `json:"submission"`
	HasParent  bool                  `json:"has_parent"`

	Blocks       []slack.Block `json:"-"`
	ReceivedView *slack.View   `json:"-"`

	inputID   int
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
			BlockSet: m.Blocks,
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

func (m *modalSlack) state() *slack.ViewState {
	if m.ReceivedView != nil {
		return m.ReceivedView.State
	}
	if m.Submission != nil {
		return m.Submission.SubmittedState
	}
	return nil
}

func (m *modalSlack) Text(message string) {
	if m == nil {
		return
	}

	m.Blocks = append(m.Blocks, slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: message,
		},
		nil,
		nil,
	))
}

func (m *modalSlack) Select(text string, options []string) string {
	defer func() {
		m.inputID++
	}()

	optionHash := hashstr(strings.Join(options, ","))

	var (
		inputBlockID     string = fmt.Sprintf("input-%v-%v", optionHash, m.inputID)
		inputSelectionID string = fmt.Sprintf("input%vselection", m.inputID)
	)

	var optionObjects []*slack.OptionBlockObject
	for index, option := range options {
		optionID := fmt.Sprintf("input%voption%v", m.inputID, index)
		optionObjects = append(
			optionObjects,
			slack.NewOptionBlockObject(
				optionID,
				slack.NewTextBlockObject(slack.PlainTextType, option, false, false),
				nil,
			),
		)
	}

	input := slack.NewInputBlock(
		inputBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, text, false, false),
		nil,
		slack.NewOptionsSelectBlockElement(
			slack.OptTypeStatic,
			slack.NewTextBlockObject(slack.PlainTextType, text, false, false),
			inputSelectionID,
			optionObjects...,
		),
	)
	input.DispatchAction = true

	m.Blocks = append(m.Blocks,
		input,
	)

	if state := m.state(); state != nil {
		viewState := state.Values
		if viewState[inputBlockID][inputSelectionID].SelectedOption.Text != nil {
			return viewState[inputBlockID][inputSelectionID].SelectedOption.Text.Text
		}
	}

	return ""
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

func (m *modalSlack) handleRequest(req *socketmode.Request, metadata []byte, hash string, client *socketmode.Client) error {
	var err error

	if m.Submission != nil {
		return m.Submission.handleRequest(req, metadata, hash, client)
	}

	modal := m.render()
	modal.PrivateMetadata = string(metadata)

	var payload interface{} = map[string]interface{}{}

	switch update := m.update; update {
	case created:
		if !m.HasParent {
			_, err = client.OpenView(m.triggerID, *modal)
			if err != nil {
				return fmt.Errorf("opening view: %w", err)
			}
		} else {
			payload = slack.NewPushViewSubmissionResponse(modal)
		}
	case action:
		_, err := client.UpdateView(
			*modal,
			m.ReceivedView.ExternalID,
			hash,
			m.ReceivedView.ID,
		)
		if err != nil {
			return fmt.Errorf("updating view: %w", err)
		}
	}

	client.Ack(*req, payload)

	return nil
}

func (m *modalSlack) populateEvent(interaction slack.InteractionType, view *slack.View) error {
	if m.Submission != nil {
		return m.Submission.populateEvent(interaction, view)
	}

	m.ReceivedView = view
	if interaction == slack.InteractionTypeBlockActions {
		m.update = action
	}
	if interaction == slack.InteractionTypeViewSubmission {
		m.Submission = &modalSubmissionSlack{
			parent: m,
		}
		m.update = submitted
	}
	if interaction == slack.InteractionTypeViewClosed {
		m.update = closed
	}

	return nil
}

type modalSubmissionSlack struct {
	SubmittedState *slack.ViewState `json:"submitted_state"`
	NextModal      *modalSlack      `json:"next_modal"`

	parent *modalSlack
}

func (m *modalSubmissionSlack) Push(title string) Modal {
	if m.NextModal != nil {
		return m.NextModal
	}

	m.SubmittedState = m.parent.ReceivedView.State

	m.NextModal = &modalSlack{
		Title:     title,
		HasParent: true,
	}
	return m.NextModal
}

func (m *modalSubmissionSlack) handleRequest(req *socketmode.Request, metadata []byte, hash string, client *socketmode.Client) error {
	if m.NextModal != nil {
		return m.NextModal.handleRequest(req, metadata, hash, client)
	}

	var payload interface{} = map[string]interface{}{}
	payload = slack.NewClearViewSubmissionResponse()
	client.Ack(*req, payload)

	return nil
}

func (m *modalSubmissionSlack) populateEvent(interaction slack.InteractionType, view *slack.View) error {
	if m.NextModal != nil {
		return m.NextModal.populateEvent(interaction, view)
	}

	return nil
}

// Get sha1 from string
func hashstr(txt string) string {
	h := sha1.New()
	h.Write([]byte(txt))
	bs := h.Sum(nil)
	sh := string(fmt.Sprintf("%x\n", bs))
	return sh
}
