package app

import (
	"errors"
	"log/slog"

	"github.com/slack-go/slack"
)

// PostMessage はメッセージを投稿します
func (b *Client) PostMessage(channelID string, msgOptions ...slack.MsgOption) (string, error) {
	var slackErr slack.SlackErrorResponse
	msgOptions = append(msgOptions,
		slack.MsgOptionUsername(b.UserName()),
		slack.MsgOptionIconEmoji(b.IconEmoji()))
	_, ts, err := b.api.PostMessage(channelID, msgOptions...)
	if err != nil {
		if errors.As(err, &slackErr) {
			for _, m := range slackErr.ResponseMetadata.Messages {
				slog.Error("PostMessage", "ErrMessage", m)
			}
			return "", err
		}
		return "", err
	}
	return ts, nil
}
