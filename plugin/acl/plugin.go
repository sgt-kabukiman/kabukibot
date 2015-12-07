package acl

import (
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type pluginStruct struct {
	plugin.BasePlugin

	bot *bot.Kabukibot
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.bot = bot
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &Worker{
		bot:     self.bot,
		channel: channel,
	}
}
