package handler

import (
	"encoding/json"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"upworkfixmux/models"
)

type UserHandler struct {
	Ch chan string
}

type Parameters struct {
	Handler   *UserHandler
	JiraIssue    jira.Issue
	CustomFields map[string]string
}

type Return struct {
	S  string
	Rv []string
	Nb []string
}

var (
	Username   = os.Getenv("JIRA_USERNAME")
	Password   = os.Getenv("JIRA_PASSWORD")
	JiraUrl    = "http://jira.viantinc.com"
	WebHookUrl = os.Getenv("SLACK_WEBHOOK")
	Channel    = os.Getenv("SLACK_CHANNEL")
	SlackUser  = "ReleaseNote"
	email      = []string{"release@viantinc.com"}
	Sc = SlackClient{
		WebHookUrl: WebHookUrl,
		UserName:   SlackUser,
		Channel:    Channel,
	}

	MainOfFunctions = map[string]func(Parameters) (Return, error){
		"DB Change Request":      DbChangeRequest,
		"Runtime Change Request": RuntimeChangeRequest,
	}

	MapOfFunctions = map[string]func(Parameters) (Return, error){
		"RuntimeDefault":         RuntimeDefault,
		"Mediator":               Mediator,
		"ETL":                    ETL,
	}
)

func DbChangeRequest(p Parameters) (Return, error) {
	tm, zn, dt := Time()
	tz := fmt.Sprintf("%s %s", tm, zn)

	message := &models.DbChangeMessage{
		Action:           "trigger",
		ReleaseType:      "db",
		UpdateDashBoard:  false,
		DeploymentTicket: p.JiraIssue.Key,
		DeploymentDate:   dt,
		StartTime:        tz,
		DeploymentType:   "db_change",
	}

	m := models.Components{
		Component: "db",
		ProdBuild: "none",
		NewBuild:  "none",
	}
	message.Components = append(message.Components, m)
	// json version of message
	log.Println(message)

	r := Return{}
	// string version
	data, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return r, err
	}
	strMessage := string(data)
	log.Println(strMessage)
	r.S = strMessage
	return r, nil
}

func RuntimeChangeRequest(p Parameters) (Return, error) {
	var rb, nv string
	var rv, nb []string
	var f func(Parameters) (Return, error)

	r := p.CustomFields["customfield_10535"]

	switch r {
	case "Mediator":
		f = MapOfFunctions["Mediator"]
	default:
		f = MapOfFunctions["RuntimeDefault"]
	}

	R, err := f(p)
	if err != nil {
		return R, err
	}

	rv = R.Rv
	nb = R.Nb

	if len(nb) != 0 || len(rv) != 0 {
		if r == "Mediator" {
			rb = rv[0]
			nv = nb[len(nb)-1]
		} else {
			rb = rv[len(rv)-1]
			nv = nb[len(nb)-1]
		}
	} else {

		// post to slack
		log.Println("empty")
	}
	// checking and waiting for response from slack for rollback value or new value ONLY IF CANT FIND FROM TICKET
	if len(rb) == 0 {
		err := TriggerSlack()
		counter := 0
		for err != nil {
			err = TriggerSlack()
			if counter > 2 {
				break
			}
			time.Sleep(2 * time.Second)
			counter++
		}
		loop := true
		// TODO this channels can receive concurrently but there can be error with received message
		// because of timeout channel is slack endpoint can be blocked and crash
		for loop {
			select {
			case res := <-p.Handler.Ch:
				if len(res) > 1 {
					fmt.Println(res)
					rb = res

				} else {
					fmt.Println("Empty message")
				}
				loop = false
				break
			case <-time.After(2 * time.Minute):
				fmt.Println("timeout 2")
				loop = false

				go func() {
					msg := <-p.Handler.Ch
					log.Println(msg)
				}()
				break
			}
		}
	}
	log.Println("Service:", r, "\nRollback:", rb, "\nNewValue:", nv)

	if len(rb) < 4 {
		//To send a notification with status (slack attachments)
		sr := SlackJobNotification{
			Text:      "Error Getting Ticket Info\n",
			Color:     "warning",
			IconEmoji: ":hammer_and_wrench",
		}

		log.Println("Rollback value not compiled expected 1.0.XXXX got: " + rb)
		sr.Details = fmt.Sprintf("%s rollback value not compiled for %s,\nPlease post payload to endpoint manually", r, p.JiraIssue.Key)
		err := Sc.SendJobNotification(sr)
		if err != nil {
			log.Println(err)
		}
		return R, err
	}
	message := &models.BidderMessage{
		Action:           "trigger",
		ReleaseType:      "runtime",
		DeploymentTicket: p.JiraIssue.Key,
		UpdateJira:       "",
		DeploymentType:   "soak",
		ToEmailList:      email,
		Components:       nil,
	}

	m := models.Components{
		Component: strings.ToLower(r),
		ProdBuild: rb,
		NewBuild:  nv,
	}

	message.Components = append(message.Components, m)

	// string version
	data, err := json.Marshal(message)
	if err != nil {
		return R, err
	}

	strMessage := string(data)
	R.S = strMessage
	return R, nil
}

func RuntimeDefault(p Parameters) (Return, error) {
	//get the rollback value
	rv := strings.Split(regexp.MustCompile(`\d{1}.\d{1}.\d{5}|\d{1}.\d{1}.\d{4}`).FindString(p.CustomFields["customfield_10526"]), ".")

	//get the new build value

	nb := strings.Split(regexp.MustCompile(`adelphic\/trunk\/\d{5}`).FindString(p.CustomFields["customfield_10525"]), "/")
	if strings.TrimSpace(strings.Join(nb, "")) == "" {
		nb = strings.Split(regexp.MustCompile(`\d{1}.\d{1}.\d{5}`).FindString(p.CustomFields["customfield_10525"]), ".")
		if strings.TrimSpace(strings.Join(nb, "")) == "" {
			for _, v := range p.JiraIssue.Fields.FixVersions {
				//nb := regexp.MustCompile(`adelphic\/trunk\/\d{5}`).FindString(t["customfield_10525"])
				nb = strings.Split(regexp.MustCompile(`adelphic\/trunk\/\d{5}`).FindString(v.Name), "/")
			}
		}
	}

	r := Return{
		S:  "",
		Rv: rv,
		Nb: nb,
	}

	return r, nil
}

func Mediator(p Parameters) (Return, error) {
	rv := strings.Split(regexp.MustCompile(`\d{4}-arm|\d{4}`).FindString(p.CustomFields["customfield_10526"]), "-")

	if len(p.JiraIssue.Fields.FixVersions) > 1 {
		log.Println("Help")
	}

	nb := strings.Split(regexp.MustCompile(`\d{4}`).FindString(p.CustomFields["customfield_10525"]), ".")
	if strings.TrimSpace(strings.Join(nb, "")) == "" {
		for _, v := range p.JiraIssue.Fields.FixVersions {
			nb = strings.Split(regexp.MustCompile(`adelphic\/mediator\/\d{4}`).FindString(v.Name), "/")

		}
	}

	r := Return{
		S:  "",
		Rv: rv,
		Nb: nb,
	}

	return r, nil
}

func ETL(p Parameters) (Return, error) {

	r := Return{}
	return r, nil
}

func (h *UserHandler) RenderJson(issues []jira.Issue, resp *jira.Response, t map[string]string) (string, error) {

	log.Println("Response Code: ", resp.StatusCode)
	r := Return{}
	var err error
	var strMessage string
	for _, i := range issues {
		switch i.Fields.Type.Name {
		case "DB Change Request":
			f := MainOfFunctions["DB Change Request"]

			r, err = f(Parameters{
				Handler:      h,
				JiraIssue:    i,
				CustomFields: t,
			})

		case "Runtime Change Request":
			f := MainOfFunctions["Runtime Change Request"]

			r, err = f(Parameters{
				Handler:      h,
				JiraIssue:    i,
				CustomFields: t,
			})

		default:
			log.Println("Error from switch default")
		}

		if err != nil {
			return "", err
		}

		strMessage = r.S
	}

	return strMessage, nil
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {

	type Msg struct {
		Message string `json:"message"`
	}

	msg := Msg{}

	msg.Message = "Usage: /tag/jira_id"

	JSON(w, 200, &models.Response{
		Data: msg,
	})
}
