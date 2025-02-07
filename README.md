## Description

Slack bot project template using socket mode for individual purpose.

It makes easy to use slack-go/slack with socket mode and fast setup new project.

## Usage example

see ./cmd/example/main.go

## Required Slack Permission sopes.

Enable the socket mode and the following permissions in the Slack app settings.

required permissions:
- `chat:write`
- `chat:write.customize`
- `chat:write.public`
- `channels:history`
- `channels:join`
- `channels:manage`
- `channels:read`
- `commands`
- `groups:history`
- `groups:read`
- `groups:write`
- `reactions:write`
- `team:read`
- `users:read`

Event Subscriptions:
- `message.channels`
- `message.groups`

Optional permissions:
- `files:write`
- `metadata.message:read`
- `reactions:write`
- `im:read`
- `mpim:read`
