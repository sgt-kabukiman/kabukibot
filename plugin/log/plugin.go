package log

// TODO: This is not terribly performant, as it uses unbuffered writes. On high-frequency
// channels, this will slow down a bit, but thankfully only the goroutines for those
// channels.

import "github.com/sgt-kabukiman/kabukibot/bot"

type logConfig struct {
	Directory string
}

type pluginStruct struct {
	config logConfig
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Name() string {
	return "LOG"
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.config = logConfig{}

	err := bot.Configuration().PluginConfig(self.Name(), &self.config)
	if err != nil {
		bot.Logger().Warn("Could not load 'log' plugin configuration: %s", err)
	}
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &worker{
		directory: self.config.Directory,
		channel:   channel.Name(),
	}
}
