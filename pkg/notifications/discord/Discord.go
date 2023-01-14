package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var DiscordWebHookURL string

func createPayload(title, description string) *DiscordWebHookPayload {
	//DiscordWebHookURL = os.Getenv("DISCORD_WEBHOOK_" + teamid)

	return &DiscordWebHookPayload{
		Username: "Recreate",
		Content:  "",
		Embeds: []Embeds{
			{
				Author: Author{
					Name: "Recreate Service",
				},
				Title:       title,
				URL:         "",
				Description: description,
				Color:       ColorGreen,
				Footer: Footer{
					Text: time.Now().Format("2006-01-02 03:04"),
				},
			},
		},
	}
}

func Send(title, description string) error {
	payload := createPayload(title, description)

	j, err := json.Marshal(payload)
	if err != nil {
		log.Println("json err:", err)
		return err
	}

	req, err := http.NewRequest("POST", DiscordWebHookURL, bytes.NewBuffer(j))
	if err != nil {
		log.Println("new request err:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("client err:", err)
		return err
	}
	if resp.StatusCode != 204 {
		log.Printf("WebHook Error %#v\n", resp)
		return fmt.Errorf("WebHook Error %v", resp)
	}
	return nil
}
