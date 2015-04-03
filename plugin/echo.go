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

func (plugin *EchoPlugin) Setup(bot *bot.Kabukibot, d *bot.Dispatcher) {
	plugin.bot    = bot
	plugin.prefix = bot.Configuration().CommandPrefix

	d.OnCommandMessage(plugin.onCommand)
}

func (plugin *EchoPlugin) onCommand(msg bot.CommandMessage) {
	cmd := msg.Command()

	if (cmd == plugin.prefix + "say" || cmd == plugin.prefix + "echo") && msg.User().IsOperator {
		plugin.bot.Say(msg.Channel(), strings.Join(msg.Args(), " "))
	}
}
