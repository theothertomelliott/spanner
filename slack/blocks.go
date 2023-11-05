package slack

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/theothertomelliott/chatframework"
)

var _ chatframework.BlockUI = &Blocks{}

type Blocks struct {
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
				state.String = action.SelectedOption.Value
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

func (b *Blocks) Text(message string) {
	if b == nil {
		return
	}

	b.addText(message)
}

func (b *Blocks) addText(message string) {
	b.blocks = append(b.blocks, slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: message,
		},
		nil,
		nil,
	))
}

func (b *Blocks) Divider() {
	if b == nil {
		return
	}

	b.blocks = append(b.blocks, slack.NewDividerBlock())
}

func (b *Blocks) TextInput(label, hint, placeholder string) string {
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

func (b *Blocks) MultilineTextInput(label, hint, placeholder string) string {
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

func (b *Blocks) addTextInput(label, hint, placeholder string, multiline bool) (string, string) {
	defer func() {
		b.inputID++
	}()

	inputBlockID := fmt.Sprintf("input-%v", b.inputID)
	inputActionID := fmt.Sprintf("input%vaction", b.inputID)

	textInput := slack.NewPlainTextInputBlockElement(
		slack.NewTextBlockObject(slack.PlainTextType, placeholder, false, false),
		inputActionID,
	)
	textInput.Multiline = multiline

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

func (b *Blocks) Select(title string, options []chatframework.SelectOption) string {
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

func (b *Blocks) addSelect(text string, options []chatframework.SelectOption) (inputBlockID string, inputActionID string) {
	defer func() {
		b.inputID++
	}()

	var values []string
	for _, option := range options {
		values = append(values, option.Value)
	}
	optionHash := hashstr(strings.Join(values, ","))

	inputBlockID = fmt.Sprintf("input-%v-%v", optionHash, b.inputID)
	inputActionID = fmt.Sprintf("input%vaction", b.inputID)

	var optionObjects []*slack.OptionBlockObject
	for _, option := range options {
		var description *slack.TextBlockObject
		if option.Description != "" {
			description = slack.NewTextBlockObject(slack.PlainTextType, option.Description, false, false)
		}
		optionObjects = append(
			optionObjects,
			slack.NewOptionBlockObject(
				option.Value,
				slack.NewTextBlockObject(slack.PlainTextType, option.Label, false, false),
				description,
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

func (m *Blocks) state() map[string]BlockState {
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
