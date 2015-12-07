package subhype

import (
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type worker struct {
	plugin.NilWorker

	dict    *bot.Dictionary
	message string
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() {
		return
	}

	if !msg.IsCommand("submsg") {
		return
	}

	if !msg.IsFromBroadcaster() && !msg.IsFromOperator() {
		return
	}

	args := msg.Arguments()

	if len(args) == 0 {
		sender.Respond("you forgot to add a message: `!submsg PogChamp, {user} just became awesome!`. {user} will be replaced with the user who subscribed. To disable notifications, just disable the plugin.")
		return
	}

	text := strings.Join(args, " ")
	key := "subhype_" + strings.TrimPrefix(msg.ChannelName(), "#") + "_message"

	self.message = text
	self.dict.Set(key, text)

	sender.Respond("the subscriber notification has been updated.")
}

func (self *worker) HandleSubscriberNotificationMessage(msg *twitch.SubscriberNotificationMessage, sender bot.Sender) {
	uname := msg.User
	message := self.message

	if len(message) == 0 {
		return
	}

	message = strings.Replace(message, "{user}", uname, -1)
	message = strings.Replace(message, "{username}", uname, -1)
	message = strings.Replace(message, "{subscriber}", uname, -1)

	sender.SendText(message)
}

func subhypeKey(channel string) string {
	return "subhype_" + strings.TrimPrefix(channel, "#") + "_message"
}
