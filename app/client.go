// Package app is a Slack bot application
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Client is a struct that contains the Slack client and configuration
type Client struct {
	api          *slack.Client
	sock         *socketmode.Client
	config       *Config
	ch           chan any
	eventHandler func(client *Client) error
	botID        string
	userID       string
	team         *slack.TeamInfo
}

// BotID returns the bot ID
func (c *Client) BotID() string {
	return c.botID
}

// UserID returns the user ID
func (c *Client) UserID() string {
	return c.userID
}

// ChannelQueueSize is the size of the channel queue
const ChannelQueueSize = 10

// NewClient creates a new Slack client
func NewClient(cfg *Config) (*Client, error) {
	var slackErr slack.SlackErrorResponse

	api := slack.New(cfg.BotToken, slack.OptionAppLevelToken(cfg.AppLevelToken))

	user, err := api.AuthTest()
	if errors.As(err, &slackErr) {
		logSlackError(&slackErr)
		return nil, fmt.Errorf("%w: AuthTest: %s", ErrSlackAPI, err)
	} else if err != nil {
		return nil, err
	}

	slog.Info("Slack", "UserID", user.UserID, "User", user.User, "BotID", user.BotID, "TeamID", user.TeamID, "Team", user.Team)

	cc, _, err := api.GetConversations(&slack.GetConversationsParameters{ExcludeArchived: true, Limit: 100, TeamID: user.TeamID})
	if errors.As(err, &slackErr) {
		logSlackError(&slackErr)
		return nil, fmt.Errorf("%w: GetConversationsParameters: %s", ErrSlackAPI, err)
	} else if err != nil {
		return nil, err
	}

	if cfg.AutoJoin {
		for _, c := range cc {
			if slices.Contains(cfg.JoinChannels, c.Name) {
				slog.Info("Slack", "JoinConversation", c.Name)
				_, _, _, err := api.JoinConversation(c.ID)
				if errors.As(err, &slackErr) {
					logSlackError(&slackErr)
				} else if err != nil {
					slog.Warn("JoinConversation", "err", err)
				}
			}
		}
	}
	return &Client{api: api, config: cfg, ch: make(chan any, ChannelQueueSize), botID: user.BotID, userID: user.UserID}, nil
}

// NewSocketMode creates a new SocketMode client
func (c *Client) newSocketMode() *socketmode.Client {
	if c.sock == nil {
		c.sock = socketmode.New(c.api)
	}
	return c.sock
}

// UserName returns the user name
func (c *Client) UserName() string {
	return c.config.User.Name
}

// IconEmoji returns the icon emoji
func (c *Client) IconEmoji() string {
	return c.config.User.IconEmoji
}

// SetEventHandler sets the event handler
func (c *Client) SetEventHandler(handler func(client *Client) error) {
	c.eventHandler = handler
}

// Events returns the events from channel
func (c *Client) Events() <-chan any {
	return c.ch
}

// Run start socketmode and handle event
func (c *Client) Run() error {
	if c.eventHandler == nil {
		return ErrEventHandleNotSet
	}
	go c.eventHandler(c)

	sock := c.newSocketMode()
	go func() {
		for ev := range sock.Events {
			switch ev.Type {
			case socketmode.EventTypeEventsAPI:
				sock.Ack(*ev.Request)
				payload, _ := ev.Data.(slackevents.EventsAPIEvent)
				slog.Debug("EventsAPIEvent", "payload", payload)
				c.ch <- &payload
			case socketmode.EventTypeSlashCommand:
				sock.Ack(*ev.Request, json.RawMessage(`{"response_type": "in_channel", "text": ""}`))
				cmd, _ := ev.Data.(slack.SlashCommand)
				slog.Debug("SlachCommand", "cmd", cmd.Command, "text", cmd.Text)
				c.ch <- &cmd
			case socketmode.EventTypeConnecting:
				slog.Info("Connecting...")
			case socketmode.EventTypeConnected:
				slog.Info("Connected.")
			case socketmode.EventTypeHello:
				slog.Info("Hello!")
			default:
				slog.Warn("Skipped Unhandled SocketEvent", "event", ev)
			}
		}
	}()
	return sock.Run()
}
