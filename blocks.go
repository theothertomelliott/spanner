package chatframework

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

var _ BlockUI = &BlocksSlack{}

type BlocksSlack struct {
	blocks      []slack.Block
	BlockStates map[string]BlockState `json:"block_state,omitempty"`
	inputID     int
}

type BlockState struct {
	String string `json:"s,omitempty"`
	Int    int    `json:"i,omitempty"`
}

func blockActionToState(in map[string]map[string]slack.BlockAction) map[string]BlockState {
	out := make(map[string]BlockState)

	for blockID, block := range in {
		state := BlockState{}
		if len(block) != 1 {
			panic("expected one block action id per block")
		}
		for _, action := range block {
			if action.SelectedOption.Value != "" {
				state.String = action.SelectedOption.Text.Text
				continue
			}
			if action.Value != "" {
				state.String = action.Value
				continue
			}
		}
		out[blockID] = state
	}

	return out
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

func (b *BlocksSlack) Divider() {
	if b == nil {
		return
	}

	b.blocks = append(b.blocks, slack.NewDividerBlock())
}

func (b *BlocksSlack) TextInput(label, hint, placeholder string) string {
	inputBlockID, _ := b.addTextInput(label, hint, placeholder, false)

	// Retrieve the text from the state
	if state := b.state(); state != nil {
		viewState := state
		if state, ok := viewState[inputBlockID]; ok {
			return state.String
		}
	}

	return ""
}

func (b *BlocksSlack) MultilineTextInput(label, hint, placeholder string) string {
	inputBlockID, _ := b.addTextInput(label, hint, placeholder, true)

	// Retrieve the text from the state
	if state := b.state(); state != nil {
		viewState := state
		if state, ok := viewState[inputBlockID]; ok {
			return state.String
		}
	}

	return ""
}

func (b *BlocksSlack) addTextInput(label, hint, placeholder string, multiline bool) (string, string) {
	defer func() {
		b.inputID++
	}()

	inputBlockID := fmt.Sprintf("input-%v", b.inputID)
	inputActionID := fmt.Sprintf("input%vaction", b.inputID)

	textInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, placeholder, false, false),
		inputActionID,
	)
	textInput.Multiline = true

	input := slack.NewInputBlock(
		inputBlockID,
		slack.NewTextBlockObject(slack.PlainTextType, label, false, false),
		slack.NewTextBlockObject(slack.PlainTextType, hint, false, false),
		textInput,
	)
	input.DispatchAction = true

	b.blocks = append(b.blocks,
		input,
	)

	return inputBlockID, inputActionID
}

func (b *BlocksSlack) Select(title string, options []string) string {
	inputBlockID, _ := b.addSelect(title, options)

	// Retrieve the selected option from the state
	if state := b.state(); state != nil {
		viewState := state
		if state, ok := viewState[inputBlockID]; ok {
			return state.String
		}
	}

	return ""
}

func (b *BlocksSlack) addSelect(text string, options []string) (inputBlockID string, inputActionID string) {
	defer func() {
		b.inputID++
	}()

	optionHash := hashstr(strings.Join(options, ","))

	inputBlockID = fmt.Sprintf("input-%v-%v", optionHash, b.inputID)
	inputActionID = fmt.Sprintf("input%vaction", b.inputID)

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
			inputActionID,
			optionObjects...,
		),
	)
	input.DispatchAction = true

	b.blocks = append(b.blocks,
		input,
	)

	return inputBlockID, inputActionID
}

func (m *BlocksSlack) state() map[string]BlockState {
	if m.BlockStates != nil {
		return m.BlockStates
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
