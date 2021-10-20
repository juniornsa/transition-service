package models

type ReceivedJson struct {
	Issue     Issue `json:"issue"`
	TimeStamp uint64 `json:"timestamp"`
}

type Issue struct {
	Id     string `json:"id"`
	Self   string `json:"self"`
	Key    string `json:"key"`
	Fields Fields
}

type Fields struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Status      Status
	Reporter    Reporter
}

type Status struct {
	Name string `json:"name"`
}

type Reporter struct {
	Name string `json:"name"`
}
