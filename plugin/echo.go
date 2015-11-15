package plugin

import (
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type EchoPlugin struct {
	operator string
}

func NewEchoPlugin() *EchoPlugin {
	return &EchoPlugin{}
}

func (plugin *EchoPlugin) Setup(bot *bot.Kabukibot) {
	plugin.operator = bot.Configuration().Operator
}

func (plugin *EchoPlugin) CreateWorker(channel string) bot.PluginWorker {
	return plugin
}

func (self *EchoPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsFrom(self.operator) && (msg.IsGlobalCommand("echo") || msg.IsGlobalCommand("say")) {
		response := strings.Join(msg.Arguments(), " ")

		if len(response) == 0 {
			response = "err... echo?"
		}

		sender.SendText(response)
	}
}
