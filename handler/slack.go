package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	"transition-service/models"
)

const DefaultSlackTimeout = 5 * time.Second

type SlackClient struct {
	WebHookUrl string
	UserName   string
	Channel    string
	TimeOut    time.Duration
}

type SimpleSlackRequest struct {
	Text      string
	IconEmoji string
}

type SlackJobNotification struct {
	Color     string
	IconEmoji string
	Details   string
	Text      string
}

type SlackMessage struct {
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color         string `json:"color,omitempty"`
	Fallback      string `json:"fallback,omitempty"`
	CallbackID    string `json:"callback_id,omitempty"`
	ID            int    `json:"id,omitempty"`
	AuthorID      string `json:"author_id,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
	AuthorSubname string `json:"author_subname,omitempty"`
	AuthorLink    string `json:"author_link,omitempty"`
	AuthorIcon    string `json:"author_icon,omitempty"`
	Title         string `json:"title,omitempty"`
	TitleLink     string `json:"title_link,omitempty"`
	Pretext       string `json:"pretext,omitempty"`
	Text          string `json:"text,omitempty"`
	ImageURL      string `json:"image_url,omitempty"`
	ThumbURL      string `json:"thumb_url,omitempty"`
	// Fields and actions are not defined.
	MarkdownIn []string    `json:"mrkdwn_in,omitempty"`
	Ts         json.Number `json:"ts,omitempty"`
}

// SendSlackNotification will post to an 'Incoming Webook' url setup in Slack Apps. It accepts
// some text and the slack channel is saved within Slack.
func (sc SlackClient) SendSlackNotification(sr SimpleSlackRequest) error {
	slackRequest := SlackMessage{
		Text:      sr.Text,
		Username:  sc.UserName,
		IconEmoji: sr.IconEmoji,
		Channel:   sc.Channel,
	}
	return sc.sendHttpRequest(slackRequest)
}

func (sc SlackClient) SendJobNotification(job SlackJobNotification) error {
	attachment := Attachment{
		Color: job.Color,
		Text:  job.Details,
		Ts:    json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	slackRequest := SlackMessage{
		Text:        job.Text,
		Username:    sc.UserName,
		IconEmoji:   job.IconEmoji,
		Channel:     sc.Channel,
		Attachments: []Attachment{attachment},
	}
	return sc.sendHttpRequest(slackRequest)
}

func (sc SlackClient) SendError(message string, options ...string) (err error) {
	return sc.funcName("danger", message, options)
}

func (sc SlackClient) SendInfo(message string, options ...string) (err error) {
	return sc.funcName("good", message, options)
}

func (sc SlackClient) SendWarning(message string, options ...string) (err error) {
	return sc.funcName("warning", message, options)
}

func (sc SlackClient) funcName(color string, message string, options []string) error {
	emoji := ":hammer_and_wrench"
	if len(options) > 0 {
		emoji = options[0]
	}
	sjn := SlackJobNotification{
		Color:     color,
		IconEmoji: emoji,
		Details:   message,
	}
	return sc.SendJobNotification(sjn)
}

func (sc SlackClient) sendHttpRequest(slackRequest SlackMessage) error {
	slackBody, err := json.Marshal(slackRequest)
	if err != nil {
		log.Println(err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, sc.WebHookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if sc.TimeOut == 0 {
		sc.TimeOut = DefaultSlackTimeout
	}
	client := &http.Client{Timeout: sc.TimeOut}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	if buf.String() != "ok" {
		return fmt.Errorf("non-ok response returned from Slack %s", buf.String())
	}
	return nil
}

func (h *UserHandler) Slack(w http.ResponseWriter, r *http.Request) {


		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}

		payload := r.Form.Get("payload")

		jsonPayload := &models.Payload{}

		decoder := json.NewDecoder(bytes.NewReader([]byte(payload)))
		if err := decoder.Decode(jsonPayload); err != nil {
			log.Println(err)
			return
		}
		log.Println(jsonPayload.ResponseURL)
		for _, v := range jsonPayload.Actions {
			log.Println(v.Text)
			log.Println(v.Value)
			log.Println(v.Type)
			log.Println(v.ActionID)
			log.Println(v.ActionTs)

			// h.Ch <- v.Value
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			c := http.Client{}

			data := struct {
				Text    string `json:"text"`
				Replace bool   `json:"replace_original"`
			}{
				Text:    "Thanks for your response",
				Replace: true,
			}

			d, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				return
			}

			req, err := http.NewRequest(http.MethodPost, jsonPayload.ResponseURL, bytes.NewBuffer(d))
			if err != nil {
				log.Println(err)
				return
			}

			req.Header.Set("Content-type", "application/json")

			resp, err := c.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			var someInterface []byte

			_, err = resp.Body.Read(someInterface)
			if err != nil {
				log.Println(err)
				return
			}

			log.Println(resp.StatusCode, string(someInterface))
			wg.Done()
		}()

		wg.Wait()
		w.WriteHeader(http.StatusOK)
		return
}

func (h *UserHandler) ResponseToSlackAction(wg *sync.WaitGroup,responseUrl, someStringAsAnswer string )  {

	// can be triggered 5 times for every 30 minutes maximum
	// message can be edited buy this response url only with speed 5 times per 30 minutes
	c := http.Client{}

	data := struct {
		Text    string `json:"text"`
		Replace bool   `json:"replace_original"`
	}{
		Text:    someStringAsAnswer,
		Replace: true,
	}

	d, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, responseUrl, bytes.NewBuffer(d))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	var someInterface []byte

	_, err = resp.Body.Read(someInterface)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(resp.StatusCode, string(someInterface))
	wg.Done()
}
