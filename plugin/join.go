package plugin

import "regexp"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type JoinPlugin struct {
	bot    *bot.Kabukibot
	mngr   *bot.ChannelManager
	prefix string
}

func NewJoinPlugin() *JoinPlugin {
	return &JoinPlugin{}
}

func (self *JoinPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot = bot
	self.mngr = bot.ChannelManager()
	self.prefix = bot.Configuration().CommandPrefix

	d.OnCommand(self.onCommand, nil)
}

func (self *JoinPlugin) onCommand(cmd bot.Command) {
	if cmd.Processed() {
		return
	}

	switch cmd.Command() {
	case self.prefix + "join":
		self.handleJoin(cmd)
	case self.prefix + "part":
		fallthrough
	case self.prefix + "leave":
		self.handlePart(cmd)
	}
}

func (self *JoinPlugin) handleJoin(cmd bot.Command) {
	args := cmd.Args()
	sentOn := cmd.Channel().Name
	sender := cmd.User().Name

	var target string

	if len(args) == 0 && self.bot.IsBot(sentOn) {
		// anyone#bot: !join
		target = sender
	} else if len(args) > 0 && isChannel(args[0]) && self.bot.IsOperator(sender) {
		// op#anywhere: !join #channel
		target = args[0]
	}

	if len(target) > 0 {
		targetName := "#" + target

		if sender == target {
			targetName = "your channel"
		}

		if !self.mngr.Joined(target) {
			self.bot.Join(twitch.NewChannel(target))
			self.bot.Respond(cmd, "I've joined "+targetName+".")
		} else {
			self.bot.Respond(cmd, "I am already in "+targetName+".")
		}
	}
}

func (self *JoinPlugin) handlePart(cmd bot.Command) {
	args := cmd.Args()
	sentOn := cmd.Channel().Name
	sender := cmd.User().Name

	var target string

	if len(args) == 0 {
		if self.bot.IsBot(sentOn) {
			// (anyone)#bot: !part
			target = sender
		} else if self.bot.IsOperator(sender) || cmd.User().IsBroadcaster {
			// [op|owner]#(anywhere): !part
			target = sentOn
		}
	} else if isChannel(args[0]) && self.bot.IsOperator(sender) {
		// op#(anywhere): !part
		target = args[0]
	}

	if len(target) > 0 {
		targetName := "#" + target

		if sender == target {
			targetName = "your channel"
		}

		if !self.mngr.Joined(target) {
			self.bot.Respond(cmd, "I am not in "+targetName+".")
		} else {
			self.bot.Part(twitch.NewChannel(target))
			self.bot.Respond(cmd, "Leaving "+targetName+" now. So long and thanks for all the FrankerZ")
		}
	}
}

var joinChannelRegex = regexp.MustCompile(`^#?([a-zA-Z0-9_]+)$`)

func isChannel(name string) bool {
	return joinChannelRegex.MatchString(name)
}
