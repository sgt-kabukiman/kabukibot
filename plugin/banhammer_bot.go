package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type BanhammerBotPlugin struct {
	channelPlugin

	bot *bot.Kabukibot
}

func NewBanhammerBotPlugin() *BanhammerBotPlugin {
	return &BanhammerBotPlugin{newChannelPlugin(), nil}
}

func (self *BanhammerBotPlugin) Key() string {
	return "banhammer_bot"
}

func (self *BanhammerBotPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot = bot
}

func (self *BanhammerBotPlugin) Load(c *twitch.Channel, bot *bot.Kabukibot, d bot.Dispatcher) {
	self.addChannelListeners(c, listenerList{d.OnTwitchMessage(self.onTwitchMessage, c)})
}

func (self *BanhammerBotPlugin) onTwitchMessage(msg twitch.TwitchMessage) {
	if msg.Command() == "clearchat" {
		args := msg.Args()

		if len(args) > 0 {
			self.bot.Respond(msg, "Notification: "+args[0])
		} else {
			self.bot.Respond(msg, "Notification: chat has been cleared")
		}
	}
}
