package plugin_control

import (
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type pluginStruct struct {
	plugin.BasePlugin

	bot     *bot.Kabukibot
	prefix  string
	plugins []bot.Plugin
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.bot = bot
	self.prefix = bot.Configuration().CommandPrefix
	self.plugins = bot.Plugins()
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &worker{
		bot:     self.bot,
		prefix:  self.prefix,
		channel: channel,
		plugins: self.plugins,
	}
}
