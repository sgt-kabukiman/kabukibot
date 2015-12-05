package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"

type PingPlugin struct {
	nilWorker // yes, the worker, we want to borrow the Enable/Disable/... functions
}

func NewPingPlugin() *PingPlugin {
	return &PingPlugin{}
}

func (self *PingPlugin) Name() string {
	return ""
}

func (self *PingPlugin) Setup(bot *bot.Kabukibot) {
}

func (self *PingPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *PingPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsFromOperator() && msg.IsGlobalCommand("ping") {
		sender.SendText("Pong!")
	}
}
