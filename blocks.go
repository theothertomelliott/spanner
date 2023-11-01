package chatframework

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

type blocksSlack struct {
	blocks  []slack.Block
	inputID int
}

func (b *blocksSlack) addText(message string) {

	b.blocks = append(b.blocks, slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: message,
		},
		nil,
		nil,
	))

}

func (b *blocksSlack) addSelect(text string, options []string) (inputBlockID string, inputSelectionID string) {
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
