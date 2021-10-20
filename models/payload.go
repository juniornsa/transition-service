package models

type Payload struct {
	Type                string      `json:"type"`
	User                User        `json:"user"`
	APIAppID            string      `json:"api_app_id"`
	Token               string      `json:"token"`
	Container           Container   `json:"container"`
	TriggerID           string      `json:"trigger_id"`
	Team                Team        `json:"team"`
	Enterprise          interface{} `json:"enterprise"`
	IsEnterpriseInstall bool        `json:"is_enterprise_install"`
	State               State       `json:"state"`
	ResponseURL         string      `json:"response_url"`
	Actions             []Actions   `json:"actions"`
}
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	TeamID   string `json:"team_id"`
}
type Container struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type Team struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}
type PlainTextInputAction struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}
type ZeroMY struct {
	PlainTextInputAction PlainTextInputAction `json:"plain_text_input-action"`
}
type Values struct {
	ZeroMY ZeroMY `json:"0MY"`
}
type State struct {
	Values Values `json:"values"`
}
type Text struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji"`
}
type Actions struct {
	Type     string `json:"type"`
	BlockID  string `json:"block_id"`
	ActionID string `json:"action_id"`
	Text     Text   `json:"text"`
	Value    string `json:"value"`
	ActionTs string `json:"action_ts"`
}
