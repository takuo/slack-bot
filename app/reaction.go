package app

import "github.com/slack-go/slack"

// AddReaction Add reaction to a message
func (b *Client) AddReaction(channelID, timestamp, name string) error {
	return b.api.AddReaction(name, slack.ItemRef{
		Timestamp: timestamp,
		Channel:   channelID,
	})
}

// RemoveReaction Remove reaction from a message
func (b *Client) RemoveReaction(channelID, timestamp, name string) error {
	return b.api.RemoveReaction(name, slack.ItemRef{
		Timestamp: timestamp,
		Channel:   channelID,
	})
}
