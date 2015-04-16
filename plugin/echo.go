package plugin

import "strings"
import "github.com/sgt-kabukiman/kabukibot/bot"

type EchoPlugin struct {
	bot    *bot.Kabukibot
	prefix string
}

func NewEchoPlugin() *EchoPlugin {
	return &EchoPlugin{}
}

func (plugin *EchoPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	plugin.bot    = bot
	plugin.prefix = bot.Configuration().CommandPrefix

	d.OnCommand(plugin.onCommand, nil)
}

func (plugin *EchoPlugin) onCommand(cmd bot.Command) {
	if cmd.Processed() { return }

	command := cmd.Command()

	if (command == plugin.prefix + "say" || command == plugin.prefix + "echo") && plugin.bot.IsOperator(cmd.User().Name) {
		plugin.bot.Say(cmd.Channel(), strings.Join(cmd.Args(), " "))
	}
}
