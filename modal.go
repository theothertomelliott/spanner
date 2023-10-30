package main

import (
	"fmt"

	"github.com/slack-go/slack"
)

type modalSlack struct {
	TriggerID string `json:"trigger_id"`

	// modal only
	Title string `json:"title"`

	Blocks        []slack.Block                           `json:"-"`
	ReceivedState map[string]map[string]slack.BlockAction `json:"-"`

	inputID int
}

func (m *modalSlack) render() *slack.ModalViewRequest {
	if m == nil {
		return nil
	}
	modal := &slack.ModalViewRequest{
		Type:  slack.ViewType("modal"),
		Title: slack.NewTextBlockObject(slack.PlainTextType, m.Title, false, false),
		Close: slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false),

		// TODO: Should be controlled by the submit option
		// It should error out with a meaningful message if there are inputs but no submit button
		Submit: slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false),

		Blocks: slack.Blocks{
			BlockSet: m.Blocks,
		},

		CallbackID: "slackFrameworkModal1", // TODO: Change this
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

	var (
		inputBlockID     string = fmt.Sprintf("input%v", m.inputID)
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

	if m.ReceivedState != nil && m.ReceivedState[inputBlockID][inputSelectionID].SelectedOption.Text != nil {
		return m.ReceivedState[inputBlockID][inputSelectionID].SelectedOption.Text.Text
	}

	// TODO: Empty options may not render
	if len(options) > 0 {
		return options[0]
	}
	return ""
}

func (m *modalSlack) Submit(text string) bool {
	panic("not imlemented")
}
