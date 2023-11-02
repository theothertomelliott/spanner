package chatframework

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

type BlocksSlack struct {
	blocks     []slack.Block
	BlockState map[string]map[string]slack.BlockAction `json:"block_state"`
	inputID    int
}

func (b *BlocksSlack) Text(message string) {
	if b == nil {
		return
	}

	b.addText(message)
}

func (b *BlocksSlack) addText(message string) {
	b.blocks = append(b.blocks, slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: message,
		},
		nil,
		nil,
	))

}

func (b *BlocksSlack) Select(title string, options []string) string {
	inputBlockID, inputSelectionID := b.addSelect(title, options)

	// Retrieve the selected option from the state
	if state := b.state(); state != nil {
		viewState := state
		if viewState[inputBlockID][inputSelectionID].SelectedOption.Text != nil {
			return viewState[inputBlockID][inputSelectionID].SelectedOption.Text.Text
		}
	}

	return ""
}

func (b *BlocksSlack) addSelect(text string, options []string) (inputBlockID string, inputSelectionID string) {
	defer func() {
		b.inputID++
	}()

	optionHash := hashstr(strings.Join(options, ","))

	inputBlockID = fmt.Sprintf("input-%v-%v", optionHash, b.inputID)
	inputSelectionID = fmt.Sprintf("input%vselection", b.inputID)

	var optionObjects []*slack.OptionBlockObject
	for index, option := range options {
		optionID := fmt.Sprintf("input%voption%v", b.inputID, index)
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

	b.blocks = append(b.blocks,
		input,
	)

	return inputBlockID, inputSelectionID
}

func (m *BlocksSlack) state() map[string]map[string]slack.BlockAction {
	if m.BlockState != nil {
		return m.BlockState
	}
	return nil
}

// Get sha1 from string
func hashstr(txt string) string {
	h := sha1.New()
	h.Write([]byte(txt))
	bs := h.Sum(nil)
	sh := string(fmt.Sprintf("%x", bs))
	return sh
}
