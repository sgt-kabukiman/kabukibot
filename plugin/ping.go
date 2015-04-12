package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"

type PingPlugin struct {
	bot    *bot.Kabukibot
	prefix string
}

func NewPingPlugin() *PingPlugin {
	return &PingPlugin{}
}

func (plugin *PingPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	plugin.bot    = bot
	plugin.prefix = bot.Configuration().CommandPrefix

	d.OnCommand(plugin.onCommand, nil)
}

func (plugin *PingPlugin) onCommand(cmd bot.Command) {
	command := cmd.Command()

	if command == plugin.prefix + "ping" && plugin.bot.IsOperator(cmd.User().Name) {
		plugin.bot.Respond(cmd, "Pong!")
	}
}
