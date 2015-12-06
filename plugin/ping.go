package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"

type PingPlugin struct {
	BasePlugin
	NilWorker
}

func NewPingPlugin() *PingPlugin {
	return &PingPlugin{}
}

func (self *PingPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *PingPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsFromOperator() && msg.IsGlobalCommand("ping") {
		sender.SendText("Pong!")
	}
}
