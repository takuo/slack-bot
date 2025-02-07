package app

import (
	"log/slog"
)

// Secret use for secret string
type Secret string

// String Stringer interface method
func (s Secret) String() string {
	return "********"
}

// LogValue slog interface method
func (s Secret) LogValue() slog.Value {
	return slog.StringValue(s.String())
}

// MarshalJSON json marshaller
func (s Secret) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

// Config Client configuration
type Config func(*Client)

// ConfigName set app name
func ConfigName(name string) func(c *Client) {
	return func(c *Client) { c.name = name }
}

// ConfigBotToken set BotToken
func ConfigBotToken(token Secret) func(c *Client) {
	return func(c *Client) { c.botToken = token }
}

// ConfigAPPLevelToken set AppLevelToken
func ConfigAPPLevelToken(token Secret) func(c *Client) {
	return func(c *Client) { c.appLevelToken = token }
}

// ConfigUserName set UserName of Bot
func ConfigUserName(name string) func(c *Client) {
	return func(c *Client) { c.userName = name }
}

// ConfigIconEmoji set Icon Emoji of Bot
func ConfigIconEmoji(emoji string) func(c *Client) {
	return func(c *Client) { c.iconEmoji = emoji }
}

// ConfigAutoJoin set AutoJoin flag
func ConfigAutoJoin(autoJoin bool) func(c *Client) {
	return func(c *Client) { c.autoJoin = autoJoin }
}

// ConfigJoinChannels set Channel list for AutoJoin
func ConfigJoinChannels(channels []string) func(c *Client) {
	return func(c *Client) { c.joinChannels = channels }
}

// ConfigLogLevel set LogLevel
//
//	string : "debug"|"info"|"warn"|"error"
func ConfigLogLevel[T string | slog.Level](level T) func(c *Client) {
	var lv slog.Level
	switch value := any(level).(type) {
	case string:
		switch value {
		case "debug":
			lv = slog.LevelDebug
		case "warn":
			lv = slog.LevelWarn
		case "error":
			lv = slog.LevelError
		default:
			lv = slog.LevelInfo
		}
	case slog.Level:
		lv = value
	}
	return func(c *Client) {
		c.logLevel = lv
	}
}

// ConfigLogFile set logfile path name
func ConfigLogFile(path string) func(c *Client) {
	return func(c *Client) {
		c.logFile = path
	}
}

// ConfigLogger set logger
func ConfigLogger(logger *slog.Logger) func(c *Client) {
	return func(c *Client) {
		c.logger = logger
	}
}

// ConfigDebug set debug flag
func ConfigDebug(debug bool) func(c *Client) {
	return func(c *Client) { c.debug = debug }
}
