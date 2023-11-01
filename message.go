package chatframework

type messageSlack struct {
	eventMetadataSlack

	TextInternal string `json:"text"`
}

func (m *messageSlack) Text() string {
	return m.TextInternal
}
