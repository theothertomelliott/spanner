package chatframework

// BlockUI allows the creation of Slack blocks in a message or modal.
type BlockUI interface {
	Text(message string)
	TextInput(label string, hint string, placeholder string) string
	MultilineTextInput(label string, hint string, placeholder string) string
	Divider()
	Select(title string, options []SelectOption) string
	Button(label string) bool
}

// SelectOption defines an option for a select block.
type SelectOption struct {
	Label       string
	Description string
	Value       string
}

// SelectOptions is a convenience function to create a set of select options from
// a list of strings.
// The strings are used as both the label and value.
// The descriptions are left empty.
func SelectOptions(options ...string) []SelectOption {
	out := make([]SelectOption, len(options))
	for i, option := range options {
		out[i] = SelectOption{
			Label: option,
			Value: option,
		}
	}
	return out
}
