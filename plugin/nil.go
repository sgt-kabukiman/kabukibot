package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"

type BasePlugin struct{}

func (p *BasePlugin) Name() string {
	return ""
}

func (p *BasePlugin) Setup(bot *bot.Kabukibot) {
}

type NilWorker struct{}

func (nw *NilWorker) Enable() {
	// do nothing
}

func (nw *NilWorker) Disable() {
	// do nothing
}

func (nw *NilWorker) Part() {
	nw.Disable()
}

func (nw *NilWorker) Shutdown() {
	nw.Disable()
}

func (nw *NilWorker) Permissions() []string {
	return []string{}
}
