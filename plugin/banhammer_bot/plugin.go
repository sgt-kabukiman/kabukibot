package banhammer_bot

import (
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type pluginStruct struct {
	plugin.BasePlugin
	plugin.NilWorker
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Name() string {
	return "banhammer_bot"
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *pluginStruct) HandleClearChatMessage(msg *twitch.ClearChatMessage, sender bot.Sender) {
	if msg.User != "" {
		sender.Respond("Notification: " + msg.User)
	} else {
		sender.Respond("Notification: chat has been cleared")
	}
}
