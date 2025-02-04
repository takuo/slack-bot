package app

import "log/slog"

type secret string

func (s secret) LogValue() slog.Value {
	return slog.StringValue("********")
}

func (s secret) String() string {
	return string(s)
}

// Config 設定
type Config struct {
	// Name Application name for logging
	Name string `yaml:"Name"`
	// BotToken Slack bot token
	BotToken secret `yaml:"BotToken"`
	// AppLevelToken application level token for socket mode
	AppLevelToken secret `yaml:"AppLevelToken"`
	// User using Post method
	User struct {
		Name      string `yaml:"Name"`
		IconEmoji string `yaml:"IconEmoji"`
	} `yaml:"User"`
	// AutoJoin allow join channels when start up
	AutoJoin bool `yaml:"AutoJoin"`
	// JoinChannels list of channels for AutoJoin
	JoinChannels []string `yaml:"JoinChannels"`
	// LogLevel default is `info`
	LogLevel string `yaml:"LogLevel"`
	// LogFile default is the `stdout``
	LogFile string `yaml:"LogFile"`
	// Enable Debug mode on slack-go/slack
	Debug bool `yaml:"Debug"`
}
