// This package implements a custom Dispatcher based on the one of the twitch package.
package bot

import "github.com/sgt-kabukiman/kabukibot/twitch"

// Kabukibot's Dispatcher, the one plugins will use.
type Dispatcher interface {
	twitch.Dispatcher

	OnCommand(f CommandHandlerFunc) twitch.Listener
	HandleCommand(cmd Command)
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
func (d *dispatcher) OnCommand(f CommandHandlerFunc) twitch.Listener {
	return d.AddListener("BOT.CMD", f)
}

// Trigger a command event and fire all registered listeners in order.
func (d *dispatcher) HandleCommand(cmd Command) {
	d.TriggerEvent("BOT.CMD", func(listener interface{}) {
		listener.(CommandHandlerFunc)(cmd)
	})
}
