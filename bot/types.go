package bot

import "github.com/sgt-kabukiman/kabukibot/twitch"

type Command interface {
	twitch.Message

	Command() string
	Args()    []string
}

type CommandHandlerFunc func(Command)

type command struct {
	twitch.Message

	cmd  string
	args []string
}

func (cmd *command) Command() string   { return cmd.cmd  }
func (cmd *command) Args()    []string { return cmd.args }

type Plugin interface {
	Setup(*Kabukibot, Dispatcher)
}

type GlobalPlugin interface {
	Plugin
}

type ChannelPlugin interface {
	Plugin

	Load(*twitch.Channel, *Kabukibot, Dispatcher)
	Unload(*twitch.Channel, *Kabukibot, Dispatcher)
}
