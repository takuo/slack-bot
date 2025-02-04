package app

import (
	"errors"

	"github.com/slack-go/slack"
)

// PostMessage はメッセージを投稿します
func (c *Client) PostMessage(channelID string, msgOptions ...slack.MsgOption) (string, error) {
	var slackErr slack.SlackErrorResponse
	msgOptions = append(msgOptions,
		slack.MsgOptionUsername(c.UserName()),
		slack.MsgOptionIconEmoji(c.IconEmoji()))
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

// PostEphemeralMessage は ephemeral メッセージを投稿します
func (c *Client) PostEphemeralMessage(channelID, userID string, msgOptions ...slack.MsgOption) (string, error) {
	var slackErr slack.SlackErrorResponse
	msgOptions = append(msgOptions,
		slack.MsgOptionUsername(c.UserName()),
		slack.MsgOptionIconEmoji(c.IconEmoji()))
	ts, err := c.api.PostEphemeral(channelID, userID, msgOptions...)
	if err != nil {
		if errors.As(err, &slackErr) {
			c.logSlackError(&slackErr)
		}
		return "", err
	}
	return ts, nil
}
