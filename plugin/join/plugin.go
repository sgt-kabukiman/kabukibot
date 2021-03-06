package join

import (
	"regexp"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type pluginStruct struct {
	plugin.BasePlugin
	plugin.NilWorker

	bot    *bot.Kabukibot
	prefix string
	home   string
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.bot = bot
	self.home = "#" + strings.ToLower(bot.BotUsername())
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *pluginStruct) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() {
		return
	}

	if msg.IsGlobalCommand("join") {
		self.handleJoin(msg, sender)
	} else if msg.IsGlobalCommand("part") || msg.IsGlobalCommand("leave") {
		self.handlePart(msg, sender)
	}
}

func (self *pluginStruct) handleJoin(msg *bot.TextMessage, sender bot.Sender) {
	args := msg.Arguments()
	sentOn := msg.Channel
	user := msg.User.Name
	toJoin := ""

	if len(args) == 0 && sentOn == self.home {
		// anyone#bot: !join
		toJoin = user
	} else if len(args) > 0 && msg.IsFromOperator() && isChannel(args[0]) {
		// op#anywhere: !join #channel
		toJoin = args[0]
	}

	toJoin = "#" + strings.TrimPrefix(strings.ToLower(toJoin), "#")

	if len(toJoin) > 1 {
		sent := self.bot.Join(toJoin)

		go func() {
			result := <-sent

			if result {
				sender.Respond("I joined " + toJoin + ".")
			}
		}()
	}
}

func (self *pluginStruct) handlePart(msg *bot.TextMessage, sender bot.Sender) {
	args := msg.Arguments()
	sentOn := msg.Channel
	user := msg.User.Name
	toLeave := ""

	if len(args) == 0 {
		if sentOn == self.home {
			// (anyone)#bot: !part
			toLeave = user
		} else if msg.IsFromOperator() || msg.IsFromBroadcaster() {
			// [op|owner]#(anywhere): !part
			toLeave = sentOn
		}
	} else if isChannel(args[0]) && msg.IsFromOperator() {
		// op#(anywhere): !part #something
		toLeave = args[0]
	}

	toLeave = "#" + strings.TrimPrefix(strings.ToLower(toLeave), "#")

	if toLeave == self.home {
		sender.Respond("I am not leaving my home, sweet home...")
	} else if len(toLeave) > 1 {
		sender.Respond("I am trying to leave " + toLeave + "...")
		self.bot.Part(toLeave)
	}
}

var joinChannelRegex = regexp.MustCompile(`^#?([a-zA-Z0-9_]+)$`)

func isChannel(name string) bool {
	return joinChannelRegex.MatchString(name)
}
