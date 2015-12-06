package plugin

import (
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type EchoPlugin struct {
	BasePlugin
	NilWorker
}

func NewEchoPlugin() *EchoPlugin {
	return &EchoPlugin{}
}

func (self *EchoPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *EchoPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsFromOperator() && (msg.IsGlobalCommand("echo") || msg.IsGlobalCommand("say")) {
		response := strings.Join(msg.Arguments(), " ")

		if len(response) == 0 {
			response = "err... echo?"
		}

		sender.SendText(response)
	}
}
