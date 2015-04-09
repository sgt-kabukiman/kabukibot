// This package implements a custom Dispatcher based on the one of the twitch package.
package bot

import "github.com/sgt-kabukiman/kabukibot/twitch"

// Kabukibot's Dispatcher, the one plugins will use.
type Dispatcher interface {
	twitch.Dispatcher

	OnCommand(CommandHandlerFunc, *twitch.Channel) *twitch.Listener
	HandleCommand(Command)
}

type dispatcher struct {
	twitch.Dispatcher
}

// Create a new Dispatcher.
func NewDispatcher() Dispatcher {
	d := twitch.NewDispatcher()

	return &dispatcher{d}
}

// Add listener for a command (a message starting with "!___").
func (d *dispatcher) OnCommand(f CommandHandlerFunc, c *twitch.Channel) *twitch.Listener {
	return d.AddListener("BOT.CMD", c, f)
}

// Trigger a command event and fire all registered listeners in order.
func (d *dispatcher) HandleCommand(cmd Command) {
	d.TriggerEvent("BOT.CMD", cmd.Channel(), func(listener interface{}) {
		listener.(CommandHandlerFunc)(cmd)
	})
}
