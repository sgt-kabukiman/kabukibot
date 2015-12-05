package plugin

import (
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type BanhammerBotPlugin struct {
	nilWorker // yes, the worker, we want to borrow the Enable/Disable/... functions
}

func NewBanhammerBotPlugin() *BanhammerBotPlugin {
	return &BanhammerBotPlugin{}
}

func (self *BanhammerBotPlugin) Name() string {
	return "banhammer_bot"
}

func (self *BanhammerBotPlugin) Setup(bot *bot.Kabukibot) {
}

func (self *BanhammerBotPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *BanhammerBotPlugin) HandleClearChatMessage(msg *twitch.ClearChatMessage, sender bot.Sender) {
	if msg.User != "" {
		sender.Respond("Notification: " + msg.User)
	} else {
		sender.Respond("Notification: chat has been cleared")
	}
}
