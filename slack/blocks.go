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
	StringSlice []string `json:"ss,omitempty"`
	String      string   `json:"s,omitempty"`
	Int         int      `json:"i,omitempty"`
}

func blockActionToState(p eventPopulation) map[string]BlockState {
	in := p.interactionCallbackEvent.BlockActionState.Values
	out := make(map[string]BlockState)

	for _, action := range p.interactionCallbackEvent.ActionCallback.BlockActions {
		if action.Type == "button" {
			state := BlockState{}
			state.String = action.Text.Text
			out[action.BlockID] = state
		}
	}

	for blockID, block := range in {
		state := BlockState{}
		if len(block) != 1 {
			panic("expected one block action id per block")
		}
		for _, action := range block {
			for _, option := range action.SelectedOptions {
				state.StringSlice = append(state.StringSlice, option.Value)
			}
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

func (b *Blocks) Header(message string) {
	if b == nil {
		return
	}

	b.blocks = append(b.blocks, slack.NewHeaderBlock(
		&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: message,
		},
	))
}

func (b *Blocks) PlainText(text string) {
	if b == nil {
		return
	}

	b.blocks = append(b.blocks, slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: text,
		},
		nil,
		nil,
	))
}

func (b *Blocks) Markdown(text string) {
	if b == nil {
		return
	}

	b.blocks = append(b.blocks, slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: text,
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

func (b *Blocks) Select(title string, options []chatframework.Option) string {
	inputBlockID := b.addSelect(title, options)

	// Retrieve the selected option from the state
	if state := b.state(); state != nil {
		viewState := state
		if state, ok := viewState[inputBlockID]; ok {
			return state.String
		}
	}

	return ""
}

func (b *Blocks) addSelect(text string, options []chatframework.Option) (inputBlockID string) {
	defer func() {
		b.inputID++
	}()

	var values []string
	for _, option := range options {
		values = append(values, option.Value)
	}
	optionHash := hashstr(strings.Join(values, ","))

	inputBlockID = fmt.Sprintf("input-%v-%v", b.inputID, optionHash)
	inputActionID := fmt.Sprintf("input%vaction", b.inputID)

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

	return inputBlockID
}

func (b *Blocks) MultipleSelect(title string, options []chatframework.Option) []string {
	inputBlockID := b.addMultipleSelect(title, options)

	// Retrieve the selected option from the state
	if state := b.state(); state != nil {
		viewState := state
		if state, ok := viewState[inputBlockID]; ok {
			return state.StringSlice
		}
	}

	return nil
}

func (b *Blocks) addMultipleSelect(text string, options []chatframework.Option) (inputBlockID string) {
	defer func() {
		b.inputID++
	}()

	var values []string
	for _, option := range options {
		values = append(values, option.Value)
	}
	optionHash := hashstr(strings.Join(values, ","))

	inputBlockID = fmt.Sprintf("input-%v-%v", b.inputID, optionHash)
	inputActionID := fmt.Sprintf("input%vaction", b.inputID)

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
		slack.NewOptionsMultiSelectBlockElement(
			slack.MultiOptTypeStatic,
			slack.NewTextBlockObject(slack.PlainTextType, text, false, false),
			inputActionID,
			optionObjects...,
		),
	)
	input.DispatchAction = true

	b.blocks = append(b.blocks,
		input,
	)

	return inputBlockID
}

func (b *Blocks) Button(label string) bool {
	defer func() {
		b.inputID++
	}()

	inputBlockID := fmt.Sprintf("input-%v", b.inputID)
	inputActionID := fmt.Sprintf("input%vaction", b.inputID)

	buttonInput := slack.NewButtonBlockElement(
		inputActionID,
		label,
		slack.NewTextBlockObject(slack.PlainTextType, label, false, false),
	)
	actions := slack.NewActionBlock(inputBlockID, buttonInput)

	b.blocks = append(b.blocks,
		actions,
	)

	// Retrieve the selected option from the state
	if state := b.state(); state != nil {
		viewState := state
		if state, ok := viewState[inputBlockID]; ok {
			return state.String == label
		}
	}

	return false
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
