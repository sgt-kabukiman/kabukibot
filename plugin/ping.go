package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"

type PingPlugin struct {
	operator string
}

func NewPingPlugin() *PingPlugin {
	return &PingPlugin{}
}

func (plugin *PingPlugin) Setup(bot *bot.Kabukibot) {
	plugin.operator = bot.Configuration().Operator
}

func (plugin *PingPlugin) CreateWorker(channel string) bot.PluginWorker {
	return plugin
}

func (self *PingPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsFrom(self.operator) && msg.IsGlobalCommand("ping") {
		sender.SendText("Pong!")
	}
}
