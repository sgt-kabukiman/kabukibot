package plugin

import "strings"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type SubHypePlugin struct {
	channelPlugin

	bot    *bot.Kabukibot
	dict   *bot.Dictionary
	prefix string
}

func NewSubHypePlugin() *SubHypePlugin {
	return &SubHypePlugin{newChannelPlugin(), nil, nil, ""}
}

func (self *SubHypePlugin) Key() string {
	return "subhype"
}

func (self *SubHypePlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot    = bot
	self.dict   = bot.Dictionary()
	self.prefix = bot.Configuration().CommandPrefix
}

func (self *SubHypePlugin) Load(c *twitch.Channel, bot *bot.Kabukibot, d bot.Dispatcher) {
	self.addChannelListeners(c, listenerList{
		d.OnCommand(self.onCommand, c),
		d.OnTwitchMessage(self.onTwitchMessage, c),
	})
}

func (self *SubHypePlugin) onCommand(cmd bot.Command) {
	if cmd.Processed() { return }

	if cmd.Command() != "submsg" {
		return
	}

	user := cmd.User()

	if !user.IsBroadcaster && !self.bot.IsOperator(user.Name) {
		return
	}

	args := cmd.Args()

	if len(args) == 0 {
		self.bot.Respond(cmd, "you forgot to add a message: `!submsg PogChamp, {user} just became awesome!`. {user} will be replaced with the user who subscribed. To disable notifications, just disable the plugin: `!" + self.prefix + "disable subhype`.")
	}

	msg := strings.Join(args, " ")
	key := "subhype_" + cmd.Channel().Name + "_message"

	self.dict.Set(key, msg)

	self.bot.Respond(cmd, "the subscriber notification has been updated.")
}

func (self *SubHypePlugin) onTwitchMessage(msg twitch.TwitchMessage) {
	if msg.Command() != "subscriber" {
		return
	}

	cname   := msg.Channel().Name
	message := self.dict.Get("subhype_" + cname + "_message")

	if len(message) == 0 {
		return
	}

	uname := msg.User().Name

	message = strings.Replace(message, "{user}", uname, -1)
	message = strings.Replace(message, "{username}", uname, -1)
	message = strings.Replace(message, "{subscriber}", uname, -1)

	self.bot.RespondToAll(msg, message)
}
