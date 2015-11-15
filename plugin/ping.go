package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"

type PingPlugin struct {
	operator string
}

func NewPingPlugin() *PingPlugin {
	return &PingPlugin{}
}

func (self *PingPlugin) Name() string {
	return ""
}

func (self *PingPlugin) Permissions() []string {
	return []string{}
}

func (self *PingPlugin) Setup(bot *bot.Kabukibot) {
	self.operator = bot.OpUsername()
}

func (self *PingPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *PingPlugin) Part() {
	// nothing to do for us
}

func (self *PingPlugin) Shutdown() {
	// nothing to do for us
}

func (self *PingPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsFrom(self.operator) && msg.IsGlobalCommand("ping") {
		sender.SendText("Pong!")
	}
}
