package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type ConsoleOutputPlugin struct {
	bot *bot.Kabukibot
	log bot.Logger
	me  string
}

func NewConsoleOutputPlugin() *ConsoleOutputPlugin {
	return &ConsoleOutputPlugin{}
}

func (self *ConsoleOutputPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot = bot
	self.log = bot.Logger()
	self.me  = bot.Configuration().Account.Username

	d.OnTextMessage(self.onText, nil)
	d.OnTwitchMessage(self.onTwitch, nil)
	d.OnResponse(self.onResponse, nil)
}

func (self* ConsoleOutputPlugin) onText(msg twitch.TextMessage) {
	user := msg.User()

	self.log.Info("[#%s] %s%s: %s", msg.Channel().Name, self.userPrefix(user), user.Name, msg.Text())
}

func (self* ConsoleOutputPlugin) onTwitch(msg twitch.TwitchMessage) {
	switch msg.Command() {
	case "clearchat":
		args := msg.Args()

		if len(args) > 0 {
			self.log.Info("[#%s] <%s has been timed out>", msg.Channel().Name, args[1])
		} else {
			self.log.Info("[#%s] <chat has been cleared>", msg.Channel().Name)
		}

	case "subscriber":
		self.log.Info("[#%s] <%s just subscribed!>", msg.Channel().Name, msg.Args()[0])
	}
}

func (self* ConsoleOutputPlugin) onResponse(r bot.Response) {
	self.log.Info("[#%s] %%%s: %s", r.Channel().Name, self.me, r.Text())
}

func (self* ConsoleOutputPlugin) userPrefix(u *twitch.User) string {
	prefix := ""

	if self.bot.IsOperator(u.Name) { prefix += "$" }
	if u.IsBroadcaster             { prefix += "&" }
	if u.IsModerator               { prefix += "@" }
	if u.IsSubscriber              { prefix += "+" }
	if u.IsTurbo                   { prefix += "~" }
	if u.IsTwitchAdmin             { prefix += "!" }
	if u.IsTwitchStaff             { prefix += "!" }

	return prefix
}
