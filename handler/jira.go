package handler

import (
	"fmt"
	"log"

	"github.com/andygrunwald/go-jira"
)

func (h *UserHandler) JiraQl(id int, sId string) (string, jira.Issue, error) {

	jiraURL := JiraUrl
	var I jira.Issue
	tp := jira.BasicAuthTransport{
		Username: Username,
		Password: Password,
	}

	Client, err := jira.NewClient(tp.Client(), jiraURL)
	if err != nil {
		log.Println(err)
		return "", I, err
	}
	log.Println(sId)
	jql := fmt.Sprintf(`id = %d`, id)

	// Running JQL query
	issues, resp, err := Client.Issue.Search(jql, nil)
	if err != nil {
		log.Println(err)
		return "", I, err
	}

	customField, resp, err := Client.Issue.GetCustomFields(sId)
	if err != nil {
		log.Println(err)
		return "", I, err
	}

	log.Printf("Call to %s\n", resp.Request.URL)
	log.Printf("Response Code: %d\n", resp.StatusCode)
	log.Println("==================================")

	for _, i := range issues {
		log.Printf("%s (%s/%s): %+v\n", i.Key, i.Fields.Type.Name, i.Fields.Priority.Name,
			i.Fields.Summary)
		I = i
	}
	log.Println("==================================")
	msg, err := h.RenderJson(issues, resp, customField)
	if err != nil {
		return "", I, err
	}
	return msg, I, nil
}
/*
 CT-9230 (UI Change Request/Normal): h1. Deployment

*Components to Deploy*
 * UI/WS - 11694
 * Campaign Maintenance Service - {color:#ff0000}+*Do Not Deploy*+{color}
 * NUI - 2.100 - [https://github.vianttech.com/adelphic/newui/actions/runs/1155]
 * Reporting Service (GBQ) - 11694
 * Reporting Service (Vertica) - 11694
 * Forecasting Service (GBQ) - 11694
 * Cube Service (GBQ) - 11694
 * PWS - 11694
 * Micro-Services - See Below
 ** For [~smilbrandt] - CT-9213

*Additional Requests*
 * [~wtseng], [~vnandakumar], [~ktummala], [~buparkar]  - Please check the schema.ddl for tonight's deployment.
 * Please deploy these changes to WalkMe, sandbox and UAT tonight.
 * Can you please flush the CDN cache on all 3 domains?

*Notes*
 * N/A

*Rollback*
 * UI/WS - 11620
 * NUI - 2.00
 * Reporting Service (GBQ) - 11620
 * Reporting Service (Vertica) - 11620
 * Forecasting Service (GBQ) - 11620
 * Cube Service (GBQ) - 11620
 * PWS - 11620
 * Micro-Services - N/A
2021/10/10 10:05:41 jira.go:49: ==================================
 */
