package bot

import "github.com/sgt-kabukiman/kabukibot/twitch"

type Plugin interface {
	Name() string
	Setup(*Kabukibot)
	CreateWorker(Channel) PluginWorker
	Permissions() []string
}

// type GlobalPlugin interface {
// 	Plugin
// }

// type ChannelPlugin interface {
// 	Plugin

// 	Key() string
// 	Permissions() []string
// 	Load(*twitch.Channel, *Kabukibot, Dispatcher)
// 	Unload(*twitch.Channel, *Kabukibot, Dispatcher)
// }

type PluginWorker interface {
	Part()
	Shutdown()
}

type pluginWorkerStruct struct {
	Plugin  Plugin
	Worker  PluginWorker
	Enabled bool
}

// these are just used to detect message types that a plugin worker wants to handle

type textMessageWorker interface {
	HandleTextMessage(*TextMessage, Sender)
}

type roomStateMessageWorker interface {
	HandleRoomStateMessage(*twitch.RoomStateMessage, Sender)
}

type clearChatMessageWorker interface {
	HandleClearChatMessage(*twitch.ClearChatMessage, Sender)
}

type subNotificationMessageWorker interface {
	HandleSubscriberNotificationMessage(*twitch.SubscriberNotificationMessage, Sender)
}
