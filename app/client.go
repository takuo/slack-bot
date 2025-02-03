// Package app is a Slack bot application
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"slices"

	errs "github.com/pkg/errors"
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
	eventHandler func(bot *Client) error
	botID        string
	userID       string
}

// BotID returns the bot ID
func (b *Client) BotID() string {
	return b.botID
}

// UserID returns the user ID
func (b *Client) UserID() string {
	return b.userID
}

// New creates a new Slack client
func New(cfg *Config) (*Client, error) {
	var slackErr slack.SlackErrorResponse

	api := slack.New(cfg.BotToken, slack.OptionAppLevelToken(cfg.AppLevelToken))

	user, err := api.AuthTest()
	if errors.As(err, &slackErr) {
		slog.Error("AuthTest", "err", slackErr.Err, "message", slackErr.ResponseMetadata.Messages)
		return nil, errs.WithStack(err)
	} else if err != nil {
		fmt.Println(debug.Stack())
		return nil, errs.WithStack(err)
	}

	slog.Info("Slack", "UserID", user.UserID, "User", user.User, "BotID", user.BotID, "TeamID", user.TeamID, "Team", user.Team)

	cc, _, err := api.GetConversations(&slack.GetConversationsParameters{ExcludeArchived: true, Limit: 100, TeamID: user.TeamID})
	if errors.As(err, &slackErr) {
		slog.Error("GetConversationsParameters", "err", slackErr.Err, "message", slackErr.ResponseMetadata.Messages)
		return nil, errs.WithStack(err)
	} else if err != nil {
		fmt.Println(debug.Stack())
		return nil, errs.WithStack(err)
	}

	if cfg.AutoJoin {
		for _, c := range cc {
			if slices.Contains(cfg.JoinChannels, c.Name) {
				slog.Info("Slack", "Join", c.Name)
				_, _, _, err := api.JoinConversation(c.ID)
				if errors.As(err, &slackErr) {
					for _, m := range slackErr.ResponseMetadata.Messages {
						slog.Error("Join", "err", slackErr.Err, "message", m)
					}
				}
			}
		}
	}
	return &Client{api: api, config: cfg, ch: make(chan any), botID: user.BotID, userID: user.UserID}, nil
}

// newSocketMode creates a new SocketMode client
func (b *Client) newSocketMode() *socketmode.Client {
	if b.sock == nil {
		b.sock = socketmode.New(b.api)
	}
	return b.sock
}

// UserName returns the user name
func (b *Client) UserName() string {
	return b.config.User.Name
}

// IconEmoji returns the icon emoji
func (b *Client) IconEmoji() string {
	return b.config.User.IconEmoji
}

// SetEventHandler sets the event handler
func (b *Client) SetEventHandler(handler func(bot *Client) error) {
	b.eventHandler = handler
}

// Events returns the events from channel
func (b *Client) Events() <-chan any {
	return b.ch
}

// Run start socketmode and handle event
func (b *Client) Run() error {
	if b.eventHandler == nil {
		return errors.New("eventHandler is not set")
	}
	go b.eventHandler(b)

	sock := b.newSocketMode()
	go func() {
		for ev := range sock.Events {
			switch ev.Type {
			case socketmode.EventTypeEventsAPI:
				slog.Debug("EventsAPIEvent")
				sock.Ack(*ev.Request)
				payload, _ := ev.Data.(slackevents.EventsAPIEvent)
				b.ch <- &payload
			case socketmode.EventTypeSlashCommand:
				sock.Ack(*ev.Request, json.RawMessage(`{"response_type": "in_channel", "text": ""}`))
				cmd, _ := ev.Data.(slack.SlashCommand)
				slog.Debug("SlachCommand", "cmd", cmd.Command, "text", cmd.Text)
				b.ch <- &cmd
			case socketmode.EventTypeConnecting:
				slog.Info("Connecting...")
				continue
			case socketmode.EventTypeConnected:
				slog.Info("Connected.")
				continue
			case socketmode.EventTypeHello:
				slog.Info("Hello!")
			default:
				slog.Error("Skipped Unhandled SocketEvent", "event", ev)
			}
		}
	}()
	return sock.Run()
}
