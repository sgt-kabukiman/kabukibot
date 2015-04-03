package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"

type CorePlugin struct {
	bot    *bot.Kabukibot
	config *bot.Configuration
	prefix string
}

func NewCorePlugin() *CorePlugin {
	return &CorePlugin{}
}

func (plugin *CorePlugin) Setup(bot *bot.Kabukibot, d *bot.Dispatcher) {
	plugin.bot    = bot
	plugin.config = bot.Configuration()

	d.OnTextMessage(plugin.onText)
	d.OnTextMessage(plugin.printLine)
}

func (plugin *CorePlugin) Load(channel *bot.Channel, bot *bot.Kabukibot, d *bot.Dispatcher) {}
func (plugin *CorePlugin) Unload(channel *bot.Channel, bot *bot.Kabukibot, d *bot.Dispatcher) {}

func (plugin* CorePlugin) onText(msg bot.TextMessage) {
	user  := msg.User()
	cn    := msg.Channel()
	state := cn.State

	user.IsBot         = plugin.config.Account.Username == user.Name
	user.IsOperator    = plugin.config.Operator == user.Name
	user.IsBroadcaster = user.Name == cn.Name
	user.IsModerator   = cn.IsModerator(user.Name)
	user.IsSubscriber  = state.Subscriber
	user.IsTurbo       = state.Turbo
	user.IsTwitchAdmin = state.Admin
	user.IsTwitchStaff = state.Staff
	user.EmoteSet      = state.EmoteSet

	state.Clear()
}

func (plugin* CorePlugin) printLine(msg bot.TextMessage) {
	println(msg.User().Prefix() + msg.User().Name + " said '" + msg.Text() + "'")
}
