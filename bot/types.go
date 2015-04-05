package bot

import "github.com/sgt-kabukiman/kabukibot/twitch"

type Plugin interface {
	Setup(*Kabukibot, *twitch.Dispatcher)
}

type GlobalPlugin interface {
	Plugin
}

type ChannelPlugin interface {
	Plugin

	Load(*twitch.Channel, *Kabukibot, *twitch.Dispatcher)
	Unload(*twitch.Channel, *Kabukibot, *twitch.Dispatcher)
}
