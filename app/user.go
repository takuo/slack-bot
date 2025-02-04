package app

import (
	"errors"

	"github.com/slack-go/slack"
)

// GetUserInfo Get user info
func (c *Client) GetUserInfo(userID string) (*slack.User, error) {
	var slackErr slack.SlackErrorResponse
	user, err := c.api.GetUserInfo(userID)
	if err != nil {
		if errors.As(err, &slackErr) {
			c.logSlackError(&slackErr)
		}
		return nil, err
	}
	return user, nil
}
