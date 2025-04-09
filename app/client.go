// Package app is a Slack bot application
package app

import (
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

type (
	// AckResponseFunc used for ack response, should returns JSONMarshallable payload or nil
	AckResponseFunc func(*socketmode.Event) any
	// MessageHandler Message event Handler
	MessageHandler func(*slackevents.MessageEvent, *Client)
	// SlashCommandHandler Slash commmand Handler
	SlashCommandHandler func(*slack.SlashCommand, *Client)
)

// Client is a struct that contains the Slack client and configuration
type Client struct {
	logger *slog.Logger
	api    *slack.Client
	sock   *socketmode.Client
	smh    *socketmode.SocketmodeHandler

	msgHandlers   []MessageHandler
	cmdHandlerMap map[string]SlashCommandHandler
	ackFunc       AckResponseFunc

	botID  string
	userID string
	team   *slack.TeamInfo

	// config
	name          string
	appLevelToken Secret
	botToken      Secret
	userName      string
	iconEmoji     string

	autoJoin     bool
	joinChannels []string
	logLevel     slog.Level
	logFile      string
	debug        bool
}

func (c *Client) ack(ev *socketmode.Event, sock *socketmode.Client) {
	if c.ackFunc != nil {
		sock.Ack(*ev.Request, c.ackFunc(ev))
	} else {
		sock.Ack(*ev.Request)
	}
}

// SetAckResponseFunc set custom Ack function
func (c *Client) SetAckResponseFunc(f AckResponseFunc) {
	c.ackFunc = f
}

// BotID returns the BotID
func (c *Client) BotID() string {
	return c.botID
}

// UserID returns the UserID of Bot
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
		name:          "SlackBot",
		logLevel:      slog.LevelInfo,
		cmdHandlerMap: make(map[string]SlashCommandHandler),
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
			w = os.Stdout
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

	cli.api = slack.New(string(cli.botToken),
		slack.OptionAppLevelToken(string(cli.appLevelToken)),
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

// SetLogger set a slog.Logger
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logger
}

// API returns the slack.Client
func (c *Client) API() *slack.Client {
	return c.api
}

// Run start socketmode and handle event
func (c *Client) Run() error {
	c.sock = c.newSocketMode()

	c.smh = socketmode.NewSocketmodeHandler(c.sock)
	// ack
	c.smh.Handle(socketmode.EventTypeEventsAPI, c.ack)
	c.smh.Handle(socketmode.EventTypeSlashCommand, c.ack)
	c.smh.Handle(socketmode.EventTypeInteractive, c.ack)

	c.smh.Handle(socketmode.EventTypeConnecting, func(_ *socketmode.Event, _ *socketmode.Client) {
		c.logger.Info("Connecting...")
	})
	c.smh.Handle(socketmode.EventTypeConnected, func(_ *socketmode.Event, _ *socketmode.Client) {
		c.logger.Info("Connected.")
	})
	c.smh.Handle(socketmode.EventTypeConnectionError, func(ev *socketmode.Event, _ *socketmode.Client) {
		c.logger.Error("Connection Error.", "event", ev)
	})
	c.smh.Handle(socketmode.EventTypeDisconnect, func(ev *socketmode.Event, _ *socketmode.Client) {
		c.logger.Warn("Disconnected.", "event", ev)
	})
	c.smh.Handle(socketmode.EventTypeHello, func(_ *socketmode.Event, _ *socketmode.Client) {
		c.logger.Info("Hello!")
	})

	// handler wrappers
	c.smh.Handle(socketmode.EventTypeEventsAPI, c.handleMessage)
	c.smh.Handle(socketmode.EventTypeSlashCommand, c.handleSlashCommand)
	// TODO: implement interaction handler

	return c.smh.RunEventLoop()
}

// AddMessageHandler add a message event handler
func (c *Client) AddMessageHandler(f MessageHandler) {
	if f == nil {
		panic("handler function should not be nil")
	}
	c.msgHandlers = append(c.msgHandlers, f)
}

// AddSlashCommandHandler register slash command handler (only one for one command)
func (c *Client) AddSlashCommandHandler(cmd string, f SlashCommandHandler) {
	if cmd == "" {
		panic("command should not be empty")
	}
	if f == nil {
		panic("handler function should not be nil")
	}
	if _, ok := c.cmdHandlerMap[cmd]; ok {
		panic(fmt.Sprintf("duplicate handler for %s", cmd))
	}
	c.cmdHandlerMap[cmd] = f
}

func (c *Client) handleMessage(ev *socketmode.Event, _ *socketmode.Client) {
	if len(c.msgHandlers) == 0 {
		return
	}
	pld, ok := ev.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return
	}
	c.logger.Debug("EventsAPIEvent", "payload", pld)
	msg, ok := pld.InnerEvent.Data.(*slackevents.MessageEvent)
	if !ok {
		return
	}
	c.logger.Debug("MessageEvent", "event", msg)
	for _, h := range c.msgHandlers {
		h(msg, c)
	}
}

func (c *Client) handleSlashCommand(ev *socketmode.Event, _ *socketmode.Client) {
	if len(c.cmdHandlerMap) == 0 {
		return
	}
	cmd, ok := ev.Data.(slack.SlashCommand)
	if !ok {
		return
	}
	c.logger.Debug("SlachCommand", "cmd", cmd.Command, "text", cmd.Text)
	if h, ok := c.cmdHandlerMap[cmd.Command]; ok {
		h(&cmd, c)
	}
}
