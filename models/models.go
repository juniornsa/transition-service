package models

type BidderMessage struct {
	Action           string   `json:"action"`
	ReleaseType      string   `json:"release_type"`
	DeploymentTicket string   `json:"deployment_ticket"`
	UpdateJira       string   `json:"update_jira"`
	DeploymentType   string   `json:"deployment_type"`
	ToEmailList      []string `json:"to_email_list"`
	Components       []struct {
		Component string `json:"component"`
		ProdBuild string `json:"prod_build"`
		NewBuild  string `json:"new_build"`
	} `json:"components"`
}

type DbChangeMessage struct {
	Action           string       `json:"action"`
	ReleaseType      string       `json:"release_type"`
	UpdateDashBoard  bool         `json:"update_dashboard"`
	DeploymentTicket string       `json:"deployment_ticket"`
	DeploymentDate   string       `json:"deployment_date"`
	StartTime        string       `json:"start_time"`
	DeploymentType   string       `json:"deployment_type"`
	Components       []Components `json:"components"`
}

type Components struct {
	Component string `json:"component"`
	ProdBuild string `json:"prod_build"`
	NewBuild  string `json:"new_build"`
}
