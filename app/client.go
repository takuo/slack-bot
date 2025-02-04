// Package app is a Slack bot application
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Client is a struct that contains the Slack client and configuration
type Client struct {
	logger       *slog.Logger
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
	var err error
	var slackErr slack.SlackErrorResponse

	var lv slog.Level
	var w io.Writer
	switch cfg.LogLevel {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	switch cfg.LogFile {
	case "", "stdout":
		w = os.Stderr
	case "stderr":
		w = os.Stderr
	case "discard", "null", "nil", "nop":
		w = io.Discard
	default:
		w, err = os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
	}
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lv})).With("AppName", cfg.Name)
	client := &Client{config: cfg, logger: logger, ch: make(chan any, ChannelQueueSize)}

	api := slack.New(cfg.BotToken.String(), slack.OptionAppLevelToken(cfg.AppLevelToken.String()))
	client.api = api

	user, err := api.AuthTest()
	if errors.As(err, &slackErr) {
		client.logSlackError(&slackErr)
		return nil, fmt.Errorf("%w: AuthTest: %s", ErrSlackAPI, err)
	} else if err != nil {
		return nil, err
	}
	authInfo := slog.Group("SlackAuthInfo",
		slog.Group("User",
			slog.String("Name", user.User),
			slog.String("UserID", user.UserID),
			slog.String("BotID", user.BotID)),
		slog.Group("Team",
			slog.String("TeamID", user.TeamID),
			slog.String("Name", user.Team),
			slog.String("URL", user.URL)))
	logger = logger.With(authInfo)
	client.botID = user.BotID
	client.userID = user.UserID

	cc, _, err := api.GetConversations(&slack.GetConversationsParameters{ExcludeArchived: true, Limit: 100, TeamID: user.TeamID})
	if errors.As(err, &slackErr) {
		client.logSlackError(&slackErr)
		return nil, fmt.Errorf("%w: GetConversationsParameters: %s", ErrSlackAPI, err)
	} else if err != nil {
		return nil, err
	}

	if cfg.AutoJoin {
		for _, c := range cc {
			if slices.Contains(cfg.JoinChannels, c.Name) {
				logger.Info("Slack", "JoinConversation", c.Name)
				_, _, _, err := api.JoinConversation(c.ID)
				if errors.As(err, &slackErr) {
					client.logSlackError(&slackErr)
				} else if err != nil {
					logger.Warn("JoinConversation", "err", err)
				}
			}
		}
	}
	return client, nil
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

// SetLogger set a slog.Logger
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logger
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
				c.logger.Debug("EventsAPIEvent", "payload", payload)
				c.ch <- &payload
			case socketmode.EventTypeSlashCommand:
				sock.Ack(*ev.Request, json.RawMessage(`{"response_type": "in_channel", "text": ""}`))
				cmd, _ := ev.Data.(slack.SlashCommand)
				c.logger.Debug("SlachCommand", "cmd", cmd.Command, "text", cmd.Text)
				c.ch <- &cmd
			case socketmode.EventTypeConnecting:
				c.logger.Info("Connecting...")
			case socketmode.EventTypeConnected:
				c.logger.Info("Connected.")
			case socketmode.EventTypeHello:
				c.logger.Info("Hello!")
			default:
				c.logger.Warn("Skipped Unhandled SocketEvent", "event", ev)
			}
		}
	}()
	return sock.Run()
}
