package plugin

import (
	"regexp"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type JoinPlugin struct {
	bot      *bot.Kabukibot
	prefix   string
	operator string
	home     string
}

func NewJoinPlugin() *JoinPlugin {
	return &JoinPlugin{}
}

func (self *JoinPlugin) Name() string {
	return ""
}

func (self *JoinPlugin) Permissions() []string {
	return []string{}
}

func (self *JoinPlugin) Setup(bot *bot.Kabukibot) {
	self.bot = bot
	self.operator = bot.OpUsername()
	self.home = "#" + strings.ToLower(bot.BotUsername())
}

func (self *JoinPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *JoinPlugin) Part() {
	// nothing to do for us
}

func (self *JoinPlugin) Shutdown() {
	// nothing to do for us
}

func (self *JoinPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsGlobalCommand("join") {
		self.handleJoin(msg, sender)
	} else if msg.IsGlobalCommand("part") || msg.IsGlobalCommand("leave") {
		self.handlePart(msg, sender)
	}
}

func (self *JoinPlugin) handleJoin(msg *bot.TextMessage, sender bot.Sender) {
	args := msg.Arguments()
	sentOn := msg.Channel
	user := msg.User.Name
	toJoin := ""

	if len(args) == 0 && sentOn == self.home {
		// anyone#bot: !join
		toJoin = user
	} else if len(args) > 0 && msg.IsFrom(self.operator) && isChannel(args[0]) {
		// op#anywhere: !join #channel
		toJoin = args[0]
	}

	toJoin = "#" + strings.TrimPrefix(strings.ToLower(toJoin), "#")

	if len(toJoin) > 1 {
		sent := self.bot.Join(toJoin)

		go func() {
			result := <-sent

			if result {
				sender.SendText("Joined " + toJoin + ".")
			}
		}()
	}
}

func (self *JoinPlugin) handlePart(msg *bot.TextMessage, sender bot.Sender) {
	args := msg.Arguments()
	sentOn := msg.Channel
	user := msg.User.Name
	toLeave := ""

	if len(args) == 0 {
		if sentOn == self.home {
			// (anyone)#bot: !part
			toLeave = user
		} else if msg.IsFrom(self.operator) || msg.IsFromBroadcaster() {
			// [op|owner]#(anywhere): !part
			toLeave = sentOn
		}
	} else if isChannel(args[0]) && msg.IsFrom(self.operator) {
		// op#(anywhere): !part #something
		toLeave = args[0]
	}

	toLeave = "#" + strings.TrimPrefix(strings.ToLower(toLeave), "#")

	if toLeave == self.home {
		sender.SendText("Not leaving my home, sweet home...")
	} else if len(toLeave) > 1 {
		sender.SendText("Trying to leave " + toLeave + "...")
		self.bot.Part(toLeave)
	}
}

var joinChannelRegex = regexp.MustCompile(`^#?([a-zA-Z0-9_]+)$`)

func isChannel(name string) bool {
	return joinChannelRegex.MatchString(name)
}
