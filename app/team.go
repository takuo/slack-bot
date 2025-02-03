package app

import (
	"errors"

	"github.com/slack-go/slack"
)

// GetTeamInfo Get team info
func (c *Client) GetTeamInfo() (*slack.TeamInfo, error) {
	if c.team != nil {
		return c.team, nil
	}
	var slackErr slack.SlackErrorResponse
	team, err := c.api.GetTeamInfo()
	if err != nil {
		if errors.As(err, &slackErr) {
			logSlackError(&slackErr)
		}
		return nil, err
	}
	c.team = team
	return team, nil
}
