package handler

import (
	"upworkfixmux/models"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func TriggerSlack() error {
	slackMessage := ReturnSlackMessage("ReleaseNote", "Error")

	c := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}

	log.Println(slackMessage)

	b, err := json.Marshal(slackMessage)
	if err != nil {
		return err
	}
	log.Println(string(b))

	req, err := http.NewRequest("POST", "slackhookurl", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	d, err := json.Marshal(resp.Body)
	if err != nil {
		return err
	}

	log.Println(resp.StatusCode, resp.ContentLength, string(d))
	return nil
}

func ReturnSlackMessage(headerText, Label string) *models.SlackMessage {
	if headerText == "" {

		log.Println("Not allowed ")
		return nil
	}
	t := &models.Elements{
		Type:  "plain_text",
		Text:  headerText,
		Emoji: true,
	}

	headerBlock := &models.Blocks{
		Type: "header",
		Text: t,
	}

	t10 := &models.Element{
		Type:     "plain_text_input",
		ActionID: "plain_text_input-action",
	}

	label := &models.Elements{
		Type:  "plain_text",
		Text:  Label,
		Emoji: true,
	}

	actionBlock := &models.Blocks{
		Type:           "input",
		DispatchAction: true,
		Element:        t10,
		Label:          label,
	}

	m := &models.SlackMessage{}
	m.Blocks = make([]*models.Blocks, 2)

	m.Blocks[0] = headerBlock
	m.Blocks[1] = actionBlock

	return m
}
