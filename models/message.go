package models

type SlackMessage struct {
	Blocks []*Blocks `json:"blocks"`
}

type Elements struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

type Element struct {
	Type     string `json:"type"`
	ActionID string `json:"action_id,omitempty"`
}

type Accessory struct {
	Type     string    `json:"type"`
	Text     *Elements `json:"text,omitempty"`
	Value    string    `json:"value,omitempty"`
	URL      string    `json:"url,omitempty"`
	ActionID string    `json:"action_id,omitempty"`
}
type Blocks struct {
	Type           string    `json:"type"`
	Text           *Elements `json:"text,omitempty"`
	DispatchAction bool      `json:"dispatch_action,omitempty"`
	Element        *Element  `json:"element,omitempty"`
	Label          *Elements `json:"label,omitempty"`
}
