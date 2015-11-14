package bot

import "github.com/sgt-kabukiman/kabukibot/twitch"

type PluginWorker interface {
	Part()
	Shutdown()
}

type pluginWorkerStruct struct {
	Worker  PluginWorker
	Enabled bool
}

// these are just used to detect message types that a plugin worker wants to handle

type textMessageWorker interface {
	HandleTextMessage(msg *twitch.TextMessage)
}

type roomStateMessageWorker interface {
	HandleRoomStateMessage(msg *twitch.RoomStateMessage)
}

type clearChatMessageWorker interface {
	HandleClearChatMessage(msg *twitch.ClearChatMessage)
}

type subNotificationMessageWorker interface {
	HandleSubscriberNotificationMessage(msg *twitch.SubscriberNotificationMessage)
}
