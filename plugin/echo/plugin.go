package echo

import (
	"strings"

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
	if msg.IsFromOperator() && (msg.IsGlobalCommand("echo") || msg.IsGlobalCommand("say")) {
		response := strings.Join(msg.Arguments(), " ")

		if len(response) == 0 {
			response = "err... echo?"
		}

		sender.SendText(response)
	}
}
