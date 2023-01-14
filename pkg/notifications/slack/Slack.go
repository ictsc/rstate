package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var SlackWebHookURL string

type SlackWebHook struct {
	Text string `json:"text"`
}

func Send(title, description string) error {
	SlackWebHookURL = os.Getenv("SLACK_WEBHOOK_URL")
	slackPayload := &SlackWebHook{
		Text: fmt.Sprintf("[%s]  \n%s", title, description),
	}
	j, err := json.Marshal(slackPayload)
	if err != nil {
		log.Println("json err:", err)
		return err
	}

	req, err := http.NewRequest("POST", SlackWebHookURL, bytes.NewBuffer(j))
	if err != nil {
		log.Printf("new request err: %v \n", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("client err: %v", err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Printf("WebHook Error %#v\n", resp)
		return fmt.Errorf("WebHook Error %v", resp)
	}
	return nil
}
