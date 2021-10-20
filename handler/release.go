package handler

import (
	"log"
	"net/http"
	"strconv"
	"transition-service/models"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func (h *UserHandler) ParseChannel(r *http.Request) chan string {

	ch := context.Get(r, "ch")

	if CHan, ok := ch.(chan string); ok {
		return CHan
	}

	return nil
}

func (h *UserHandler) Release(w http.ResponseWriter, r *http.Request) {

	// mux.Vars to receive variables from mux url
	vars := mux.Vars(r)
	jiraId := vars["jira_id"]
	log.Println("Received Jira Id from endpoint /tag/", jiraId)

	ji, err := strconv.ParseInt(jiraId, 10, 64)
	if err != nil {
		JSON(w, 400, &models.Response{
			Error:   true,
			Message: "Error --> " + err.Error(),
		})
		return
	}

	message, issue, err := h.JiraQl(int(ji), jiraId)
	if err != nil {
		JSON(w, 400, &models.Response{
			Error:   true,
			Message: "Bad request --> " + err.Error(),
		})
		return
	}

	if message != "" {
		RestRequest(w, message, issue)
	}

	return
}
