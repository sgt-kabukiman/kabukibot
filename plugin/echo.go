package plugin

import (
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type EchoPlugin struct {
	nilWorker // yes, the worker, we want to borrow the Enable/Disable/... functions
}

func NewEchoPlugin() *EchoPlugin {
	return &EchoPlugin{}
}

func (self *EchoPlugin) Name() string {
	return ""
}

func (self *EchoPlugin) Setup(bot *bot.Kabukibot) {
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
