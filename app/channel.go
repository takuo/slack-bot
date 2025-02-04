package app

import (
	"errors"

	"github.com/slack-go/slack"
)

// LeaveConversation Leave conversation
func (c *Client) LeaveConversation(channelID string) error {
	var slackErr slack.SlackErrorResponse
	if _, err := c.api.LeaveConversation(channelID); err != nil {
		if errors.As(err, &slackErr) {
			c.logSlackError(&slackErr)
		}
		return err
	}
	return nil
}
