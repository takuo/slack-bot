package app

import (
	"errors"

	"github.com/slack-go/slack"
)

// AddReaction Add reaction to a message
func (c *Client) AddReaction(channelID, timestamp, name string) error {
	return c.api.AddReaction(name, slack.ItemRef{
		Timestamp: timestamp,
		Channel:   channelID,
	})
}

// RemoveReaction Remove reaction from a message
func (c *Client) RemoveReaction(channelID, timestamp, name string) error {
	var slackErr slack.SlackErrorResponse
	if err := c.api.RemoveReaction(name, slack.ItemRef{
		Timestamp: timestamp,
		Channel:   channelID,
	}); err != nil {
		if errors.As(err, &slackErr) {
			c.logSlackError(&slackErr)
		}
		return err
	}
	return nil
}
