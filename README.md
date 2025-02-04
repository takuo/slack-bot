## Description

slack bot project template using socket mode for individual purpose.

## Usage example

example app main code with kong and kong-yaml.

```go
package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/takuo/slack-bot/app"
)


type Example struct {
	AppConfig app.Config `type:"yamlfile" default:"config.yml" help:"App configuration file"`
	Debug     *bool      `short:"d" default:"false" help:"enable debug mode"`
	LogFile   string     `type:"file" default:"stdout" help:"filename to write log"`
	LogLevel  *string    `enum:"debug,info,warn,error" help:"Set log level"`
}

func main() {
	e := example{}
	ctx := kong.Parse(&e, kong.NamedMapper(`yamlfile`, kongyaml.YAMLFileMapper))
	if err := ctx.Run(); err != nil {
		slog.Error("Run", "error", err)
		os.Exit(1)
	}
}

func (e *Example) Run() error {
	slog.Debug("Run", "AppConfig", e.AppConfig)

	client, err := app.NewClient(&e.AppConfig)
	if err != nil {
		return err
	}
	client.SetEventHandler(handleEvent)
	return client.Run()
}

func handleEvent(client *app.Client) error {
	for ev := range client.Events() {
		switch ev := ev.(type) {
		case *slackevents.EventsAPIEvent:
			if msg, ok := ev.InnerEvent.Data.(*slackevents.MessageEvent); ok {
				if msg.User == client.UserID() || msg.BotID == client.BotID() {
                    // ignore self activity
					continue
				}
				// TODO: implement
				if err := handleSlackMessageEvent(client, msg); err != nil {
					slog.Error("handleSlackMessageEvent", "error", err)
				}
			}
		case *slack.SlashCommand:
			// TODO: implement
			if err := handleSlashCommand(client, ev); err != nil {
				slog.Error("handleSlashCommand", "error", err)
			}
		}
	}
	return nil
}
```

## Required permissions

Enable the socket mode and the following permissions in the Slack app settings.

required permissions:
- `channels:join`
- `chat:write`
- `chat:write.customize`
- `chat:write.public`
- `channels:manage`
- `channels:history`
- `channels:read`
- `commands`
- `groups:read`
- `groups:write`
- `groups:history`
- `reactions:write`
- `team:read`
- `users:read`
- `users:write`

Event Subscriptions:
- `message.channels`
- `message.groups`

Optional permissions:
- `metadata.message:read`
- `emoji:read`

