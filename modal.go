package chatframework

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

type modalSlack struct {
	Title string `json:"title"`

	Blocks       []slack.Block `json:"-"`
	ReceivedView *slack.View   `json:"-"`

	inputID int

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

	if m.ReceivedView != nil {
		viewState := m.ReceivedView.State.Values
		if viewState[inputBlockID][inputSelectionID].SelectedOption.Text != nil {
			return viewState[inputBlockID][inputSelectionID].SelectedOption.Text.Text
		}
	}

	return ""
}

func (m *modalSlack) Submit(text string) bool {
	m.submitText = &text
	return m.update == submitted
}

func (m *modalSlack) Close(text string) bool {
	m.closeText = &text
	return m.update == closed
}

// Get sha1 from string
func hashstr(txt string) string {
	h := sha1.New()
	h.Write([]byte(txt))
	bs := h.Sum(nil)
	sh := string(fmt.Sprintf("%x\n", bs))
	return sh
}
