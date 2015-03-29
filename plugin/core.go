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

func (plugin* CorePlugin) onText(msg bot.TextMessage) {
	// var
	// 	user     = message.getUser(),
	// 	channel  = message.getChannel(),
	// 	username = user.getName(),
	// 	mngr     = this.userMngr;

	user := msg.User()
	// channel := msg.Channel()

	user.IsBot         = plugin.config.Account.Username == user.Name
	user.IsOperator    = plugin.config.Operator == user.Name
	user.IsSubscriber  = false
	user.IsTurbo       = false
	user.IsTwitchAdmin = false
	user.IsTwitchStaff = false

	// user.setSubscriber(mngr.takeSubscriber(channel, username));
	// user.setTurbo(mngr.takeTurboUser(username));
	// user.setTwitchStaff(mngr.isTwitchStaff(username));
	// user.setTwitchAdmin(mngr.isTwitchAdmin(username));
	// user.setEmoteSets(mngr.takeEmoteSets(username));
	// user.setOperator(mngr.isOperator(username));
	// user.setBot(username === this.botName);
}

func (plugin* CorePlugin) printLine(msg bot.TextMessage) {
	println(msg.User().Prefix() + msg.User().Name + " said '" + msg.Text() + "'")
}
