package plugin

import (
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type BanhammerBotPlugin struct {
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

func (self *BanhammerBotPlugin) Enable() {
	// nothing to do for us
}

func (self *BanhammerBotPlugin) Disable() {
	// nothing to do for us
}

func (self *BanhammerBotPlugin) Part() {
	// nothing to do for us
}

func (self *BanhammerBotPlugin) Shutdown() {
	// nothing to do for us
}

func (self *BanhammerBotPlugin) Permissions() []string {
	return []string{}
}

func (self *BanhammerBotPlugin) HandleClearChatMessage(msg *twitch.ClearChatMessage, sender bot.Sender) {
	if msg.User != "" {
		sender.Respond("Notification: " + msg.User)
	} else {
		sender.Respond("Notification: chat has been cleared")
	}
}
