## Description

slack bot project template using socket mode for individual purpose.

## Usage example

example app main code with kong and kong-yaml.

```go
package main

import (
	"log/slog"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/takuo/slack-bot/app"
)

type Example struct {
	client *app.Client
}

func main() {
	e := &Example{}
	e.client, err := app.NewClient("Example",
        app.ConfigAPPLevelToken("xapp-1-XXXXXX"),
        app.ConfigBotToken("xoxb-XXXXXXX"),
	)
	e.SetEventHandler(e.handleEvent)
	return e.client.Run()
}

func (e *Example) handleEvent(c *app.Client, ev any) error {
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
	// if returns non-nil error, sockemode will be disconnected.
	return nil
}
```

## Required Slack Permission sopes.

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

