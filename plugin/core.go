package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"

type CorePlugin struct{}

func NewCorePlugin() *CorePlugin {
	return &CorePlugin{}
}

func (plugin *CorePlugin) Setup(bot *bot.Kabukibot, d *bot.Dispatcher) {

}
