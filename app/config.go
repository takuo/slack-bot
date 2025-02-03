package app

// Config 設定
type Config struct {
	BotToken       string `yaml:"BotToken"`
	AppLevelToken  string `yaml:"AppLevelToken"`
	EnableReply    bool   `yaml:"EnableReply"`
	EnableReaction bool   `yaml:"EnableReaction"`
	User           struct {
		Name      string `yaml:"Name"`
		IconEmoji string `yaml:"IconEmoji"`
	} `yaml:"User"`
	AutoJoin     bool     `yaml:"AutoJoin"`
	JoinChannels []string `yaml:"JoinChannels"`
	LogLevel     string   `yaml:"LogLevel"`
	Debug        bool     `yaml:"Debug"`
}
