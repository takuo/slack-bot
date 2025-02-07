// Package main is example application
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/takuo/slack-bot/app"
)

// Example example app
type Example struct {
	client *app.Client
}

func main() {
	var err error
	e := &Example{}
	e.client, err = app.NewClient("Example",
		app.ConfigLogLevel(slog.LevelDebug),
		app.ConfigDebug(false),
		app.ConfigAPPLevelToken(app.Secret(os.Getenv("SLACK_APP_LEVEL_TOKEN"))),
		app.ConfigBotToken(app.Secret(os.Getenv("SLACK_BOT_TOKEN"))),
	)
	if err != nil {
		panic(err)
	}
	e.client.AddMessageHandler(e.MessageHandler)
	e.client.AddSlashCommandHandler("/example1", e.ExampleHandler1)
	e.client.AddSlashCommandHandler("/example2", e.ExampleHandler2)
	e.client.SetAckFunc(func(ev *socketmode.Event, c *app.Client) {
		switch ev.Type {
		case socketmode.EventTypeSlashCommand:
			// display command as user's message
			c.Ack(*ev.Request, json.RawMessage(`{"response_type": "in_channel", "text": ""}`))
		default:
			app.DefaultAckFunc(ev, c)
		}
	})
	if err := e.client.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// MessageHandler handle any user message
func (e *Example) MessageHandler(msg *slackevents.MessageEvent, _ *app.Client) {
	if msg.User == e.client.UserID() || msg.BotID == e.client.BotID() {
		// ignore self activity
		return
	}
	slog.Info("Message", "text", msg.Text)
}

// ExampleHandler1 handle slash command `/example1`
func (e *Example) ExampleHandler1(cmd *slack.SlashCommand, _ *app.Client) {
	slog.Info("SlashCommand", "command", cmd.Command, "text", cmd.Text)
}

// ExampleHandler2 handle slash command `/example2`
func (e *Example) ExampleHandler2(cmd *slack.SlashCommand, _ *app.Client) {
	slog.Info("SlashCommand", "command", cmd.Command, "text", cmd.Text)
}
