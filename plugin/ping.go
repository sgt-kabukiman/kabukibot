package plugin

import (
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type PingPlugin struct {
	operator string
	prefix   string
}

func NewPingPlugin() *PingPlugin {
	return &PingPlugin{}
}

func (plugin *PingPlugin) Setup(bot *bot.Kabukibot) {
	plugin.operator = bot.Configuration().Operator
	plugin.prefix = bot.Configuration().CommandPrefix
}

func (plugin *PingPlugin) CreateWorker(channel string) bot.PluginWorker {
	return plugin
}

func (self *PingPlugin) HandleTextMessage(msg *twitch.TextMessage, sender bot.Sender) {
	if strings.ToLower(msg.User.Name) == self.operator && strings.HasPrefix(msg.Text, "!"+self.prefix+"ping") {
		sender.SendText("Pong!")
	}
}
