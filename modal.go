package chatframework

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type modalSlack struct {
	Title string `json:"title"`

	Blocks         []slack.Block    `json:"-"`
	ReceivedView   *slack.View      `json:"-"`
	SubmittedState *slack.ViewState `json:"submitted_state"`

	parentModal *modalSlack
	NextModal   *modalSlack `json:"next_modal"`

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
	return m.SubmittedState
}

func (m *modalSlack) Text(message string) {
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

func (m *modalSlack) Submit(text string) bool {
	m.submitText = &text

	// Definitely submitted if a new modal has been pushed
	if m.NextModal != nil {
		return true
	}

	return m.update == submitted
}

func (m *modalSlack) Close(text string) bool {
	m.closeText = &text
	return m.update == closed
}

func (m *modalSlack) Push(title string) Modal {
	if m == nil {
		return nil
	}

	if m.NextModal != nil {
		return m.NextModal
	}

	m.SubmittedState = m.ReceivedView.State

	m.NextModal = &modalSlack{
		Title:       title,
		parentModal: m,
	}
	return m.NextModal
}

func (m *modalSlack) handleRequest(req *socketmode.Request, es *eventSlack, client *socketmode.Client) error {
	if m.NextModal != nil {
		return m.NextModal.handleRequest(req, es, client)
	}

	modal := m.render()
	metadata, err := json.Marshal(es.state)
	if err != nil {
		log.Printf("saving metadata: %v\n", err)
	}
	modal.PrivateMetadata = string(metadata)

	var payload interface{} = map[string]interface{}{}

	switch update := m.update; update {
	case created:
		if m.parentModal == nil {
			_, err = client.OpenView(m.triggerID, *modal)
			if err != nil {
				return fmt.Errorf("opening view: %w", err)
			}
		} else if m.parentModal.update == submitted {
			payload = slack.NewPushViewSubmissionResponse(modal)
		} else {
			_, err = client.PushView(m.triggerID, *modal)
			if err != nil {
				return fmt.Errorf("opening view: %w", err)
			}
		}
	case action:
		_, err := client.UpdateView(
			*modal,
			m.ReceivedView.ExternalID,
			es.hash,
			m.ReceivedView.ID,
		)
		if err != nil {
			return fmt.Errorf("updating view: %w", err)
		}
	case submitted:
		payload = slack.NewClearViewSubmissionResponse()
	}

	client.Ack(*req, payload)

	return nil
}

func (m *modalSlack) populateEvent(interaction slack.InteractionType, view *slack.View) error {
	if m.NextModal != nil {
		return m.NextModal.populateEvent(interaction, view)
	}

	m.ReceivedView = view
	if interaction == slack.InteractionTypeBlockActions {
		m.update = action
	}
	if interaction == slack.InteractionTypeViewSubmission {
		m.update = submitted
	}
	if interaction == slack.InteractionTypeViewClosed {
		m.update = closed
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
