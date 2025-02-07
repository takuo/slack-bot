// Package app is a Slack bot application
package app

import (
	"context"
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
	"github.com/sourcegraph/conc/pool"
)

// Client is a struct that contains the Slack client and configuration
type Client struct {
	logger       *slog.Logger
	api          *slack.Client
	sock         *socketmode.Client
	ch           chan any
	eventHandler func(*Client, any) error
	botID        string
	userID       string
	team         *slack.TeamInfo

	// config
	name          string
	appLevelToken secret
	botToken      secret
	userName      string
	iconEmoji     string

	autoJoin     bool
	joinChannels []string
	logLevel     slog.Level
	logFile      string
	debug        bool
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
// Required scopes: `channels:read`, `groups:read`, `im:read`, `mpim:read`, `channels:join`
func NewClient(name string, configs ...Config) (*Client, error) {
	var err error
	var slackErr slack.SlackErrorResponse
	cli := &Client{
		name:     "SlackBot",
		logLevel: slog.LevelInfo,
		ch:       make(chan any, ChannelQueueSize),
	}

	for _, cfg := range configs {
		cfg(cli)
	}

	if cli.debug {
		cli.logLevel = slog.LevelDebug
	}

	if cli.logger == nil {
		var w io.Writer
		switch cli.logFile {
		case "", "stdout":
			w = os.Stderr
		case "stderr":
			w = os.Stderr
		case "discard", "null", "nil", "nop":
			w = io.Discard
		default:
			w, err = os.OpenFile(cli.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
		}
		cli.logger = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: cli.logLevel})).With("AppName", name)
	}

	cli.api = slack.New(cli.botToken.String(),
		slack.OptionAppLevelToken(cli.appLevelToken.String()),
		slack.OptionLog(&logger{log: cli.logger, prefix: "api: "}),
		slack.OptionDebug(cli.debug))

	user, err := cli.api.AuthTest()
	if errors.As(err, &slackErr) {
		cli.logSlackError(&slackErr)
		return nil, fmt.Errorf("%w: AuthTest: %s", ErrSlackAPI, err)
	} else if err != nil {
		return nil, err
	}
	if cli.debug {
		authInfo := slog.Group("SlackAuthInfo",
			slog.Group("User",
				slog.String("Name", user.User),
				slog.String("UserID", user.UserID),
				slog.String("BotID", user.BotID)),
			slog.Group("Team",
				slog.String("TeamID", user.TeamID),
				slog.String("Name", user.Team),
				slog.String("URL", user.URL)))
		cli.logger = cli.logger.With(authInfo)
	}
	cli.botID = user.BotID
	cli.userID = user.UserID

	channels, _, err := cli.api.GetConversations(&slack.GetConversationsParameters{ExcludeArchived: true, Limit: 100, TeamID: user.TeamID})
	if errors.As(err, &slackErr) {
		cli.logSlackError(&slackErr)
		return nil, fmt.Errorf("%w: GetConversationsParameters: %s", ErrSlackAPI, err)
	} else if err != nil {
		return nil, err
	}

	if cli.autoJoin {
		for _, ch := range channels {
			if slices.Contains(cli.joinChannels, ch.Name) {
				cli.logger.Info("Slack", "JoinConversation", ch.Name)
				_, _, _, err := cli.api.JoinConversation(ch.ID)
				if errors.As(err, &slackErr) {
					cli.logSlackError(&slackErr)
				} else if err != nil {
					cli.logger.Warn("JoinConversation", "err", err)
				}
			}
		}
	}
	return cli, nil
}

// NewSocketMode creates a new SocketMode client
// Required scopes: `connections:write`
func (c *Client) newSocketMode() *socketmode.Client {
	if c.sock == nil {
		c.sock = socketmode.New(c.api,
			socketmode.OptionDebug(c.debug),
			socketmode.OptionLog(&logger{log: c.logger, prefix: "sock: "}),
		)
	}
	return c.sock
}

// UserName returns the user name
func (c *Client) UserName() string {
	return c.userName
}

// IconEmoji returns the icon emoji
func (c *Client) IconEmoji() string {
	return c.iconEmoji
}

// SetEventHandler sets the event handler
func (c *Client) SetEventHandler(handler func(*Client, any) error) {
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
	p := pool.New().WithContext(context.Background()).WithCancelOnError()
	if c.eventHandler == nil {
		return ErrEventHandleNotSet
	}
	p.Go(func(ctx context.Context) error {
		for {
			select {
			case ev := <-c.ch:
				if err := c.eventHandler(c, ev); err != nil {
					return err
				}
			case <-ctx.Done():
				slog.Warn("EventHandlerLoop", "ctx", "done")
				return nil
			}
		}
	})

	sock := c.newSocketMode()
	p.Go(func(ctx context.Context) error {
		for {
			select {
			case ev := <-sock.Events:
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
				case socketmode.EventTypeDisconnect:
					c.logger.Warn("Disconnected", "ev", ev)
				default:
					c.logger.Warn("Skipped Unhandled SocketEvent", "event", ev)
				}
			case <-ctx.Done():
				slog.Warn("SocketEventLoop", "ctx", "done")
				return nil
			}
		}
	})
	p.Go(func(ctx context.Context) error { return sock.RunContext(ctx) })
	return p.Wait()
}
