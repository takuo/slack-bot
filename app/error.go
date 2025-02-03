package app

import (
	"errors"
	"log/slog"

	"github.com/slack-go/slack"
)

var (
	// ErrSlackAPI is an error that occurs when the Slack API returns an error
	ErrSlackAPI = errors.New("SlackAPIError")
	// ErrEventHandleNotSet is an error that occurs when the event handler is not set
	ErrEventHandleNotSet = errors.New("EventHandleNotSet")
)

// logSlackError logging Slack API error
func logSlackError(err *slack.SlackErrorResponse) {
	for _, m := range err.ResponseMetadata.Messages {
		slog.Error("SlackAPIError", "message", m)
	}
}
