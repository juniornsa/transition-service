package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"upworkfixmux/models"

	"github.com/andygrunwald/go-jira"
)

var Demo bool = true

func RestRequest(w http.ResponseWriter, s string, issue jira.Issue) {

	sr := SlackJobNotification{
		Text:      "*Error Sending Payload*\n",
		Color:     "danger",
		IconEmoji: ":hammer_and_wrench",
	}
	sk := SlackJobNotification{
		Text:      "*Success*\n",
		Color:     "good",
		IconEmoji: ":perfect:",
	}

	if !Demo {
		client := &http.Client{}
		response := bytes.NewBuffer([]byte(s))
		req, err := http.NewRequest("POST",
			"https://hhkun6d0i4-vpce-0dc63ab1993c84223.execute-api.us-east-1.amazonaws.com/v1/",
			response)
		if err != nil {
			log.Println(err)
			return
		}

		//req.Header.Set("Content-type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Println("error in send req: ", err)
			w.WriteHeader(400)
			JSON(w, 400, &models.Response{
				Error:   true,
				Message: "Error message",
				Data:    nil,
			})
			return
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Println(err)
			}
		}()

		log.Println(resp.StatusCode)
		if resp.StatusCode >= 299 {
			sr.Details = fmt.Sprintf("ReleaseNote lambda error %s,\nPlease check lambda and post payload to endpoint manually\nERROR %d", issue.Key, resp.StatusCode)
			err := Sc.SendJobNotification(sr)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			sk.Details = fmt.Sprintf("*Ticket Number:* %s\n*Deployment Type*: %s", issue.Key, issue.Fields.Type.Name)
			err := Sc.SendJobNotification(sk)
			if err != nil {
				log.Fatal(err)
			}
		}
		//log.Println(resp.Body)
		var data models.ResponseFromApi
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&data); err != nil {
			log.Println(err)
			JSON(w, 400, &models.Response{
				Error:   true,
				Message: "Error",
				Data:    data,
			})
			return
		}

		JSON(w, 200, &models.Response{
			Error:   false,
			Message: "No error. Success.",
			Data:    data,
		})
	}
	return
}
