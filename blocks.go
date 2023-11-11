package spanner

// BlockUI allows the creation of Slack blocks in a message or modal.
type BlockUI interface {
	Header(message string)
	PlainText(text string)
	Markdown(text string)
	TextInput(label string, hint string, placeholder string) string
	MultilineTextInput(label string, hint string, placeholder string) string
	Divider()
	Select(title string, options []Option) string
	MultipleSelect(title string, options []Option) []string
	Button(label string) bool
}

// Option defines an option for select or checkbox blocks.
type Option struct {
	Label       string
	Description string
	Value       string
}

// Options is a convenience function to create a set of options from
// a list of strings.
// The strings are used as both the label and value.
// The descriptions are left empty.
func Options(options ...string) []Option {
	out := make([]Option, len(options))
	for i, option := range options {
		out[i] = Option{
			Label: option,
			Value: option,
		}
	}
	return out
}
