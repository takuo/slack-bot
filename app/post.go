package app

import (
	"errors"

	"github.com/slack-go/slack"
)

// PostMessage post a message
// Required scopes: `chat:write`
func (c *Client) PostMessage(channelID string, msgOptions ...slack.MsgOption) (string, error) {
	var slackErr slack.SlackErrorResponse
	if c.UserName() != "" {
		msgOptions = append(msgOptions,
			slack.MsgOptionUsername(c.UserName()))
	}
	if c.IconEmoji() != "" {
		msgOptions = append(msgOptions,
			slack.MsgOptionIconEmoji(c.IconEmoji()))
	}
	_, ts, err := c.api.PostMessage(channelID, msgOptions...)
	if err != nil {
		if errors.As(err, &slackErr) {
			c.logSlackError(&slackErr)
			return "", err
		}
		return "", err
	}
	return ts, nil
}

// PostEphemeralMessage post an ephemeral message
// Required scopes: `chat:write`
func (c *Client) PostEphemeralMessage(channelID, userID string, msgOptions ...slack.MsgOption) (string, error) {
	var slackErr slack.SlackErrorResponse
	if c.UserName() != "" {
		msgOptions = append(msgOptions,
			slack.MsgOptionUsername(c.UserName()))
	}
	if c.IconEmoji() != "" {
		msgOptions = append(msgOptions,
			slack.MsgOptionIconEmoji(c.IconEmoji()))
	}
	ts, err := c.api.PostEphemeral(channelID, userID, msgOptions...)
	if err != nil {
		if errors.As(err, &slackErr) {
			c.logSlackError(&slackErr)
		}
		return "", err
	}
	return ts, nil
}
