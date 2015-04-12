package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type listenerList   []*twitch.Listener
type channelKeysMap map[string]listenerList

type channelPlugin struct {
	listeners channelKeysMap
}

func newChannelPlugin() channelPlugin {
	return channelPlugin{make(channelKeysMap)}
}

func (self *channelPlugin) Unload(c *twitch.Channel, bot *bot.Kabukibot, d bot.Dispatcher) {
	self.removeChannelListeners(c)
}

func (self *channelPlugin) addChannelListeners(c *twitch.Channel, list listenerList) {
	self.listeners[c.Name] = list
}

func (self *channelPlugin) removeChannelListeners(c *twitch.Channel) {
	list, exists := self.listeners[c.Name]
	if exists {
		for _, listener := range list {
			listener.Remove()
		}

		delete(self.listeners, c.Name)
	}
}
