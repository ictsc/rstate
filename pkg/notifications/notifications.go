package notifications

import (
	"github.com/ictsc/rstate/pkg/notifications/discord"
	"github.com/ictsc/rstate/pkg/notifications/slack"
)

type Notifications struct {
	title       string
	description string
	channel     string
}

func NewNotifications(title, description, channel string) *Notifications {
	return &Notifications{
		title:       title,
		description: description,
		channel:     channel,
	}
}

func (c *Notifications) SendAll() int {
	result := 0
	if err := c.SendSlack(); err != nil {
		result++
	}
	if err := c.SendDiscord(); err != nil {
		result++
	}
	return result
}

func (c *Notifications) SendDiscord() error {
	return discord.Send(c.title, c.description)
}

func (c *Notifications) SendSlack() error {
	return slack.Send(c.title, c.description, c.channel)
}
