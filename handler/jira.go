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
