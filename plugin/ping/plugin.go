package ping

import (
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type pluginStruct struct {
	plugin.BasePlugin
	plugin.NilWorker
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *pluginStruct) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsFromOperator() && msg.IsGlobalCommand("ping") {
		sender.SendText("Pong!")
	}
}
