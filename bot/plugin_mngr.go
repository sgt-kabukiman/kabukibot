package bot

import "log"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type pluginList []Plugin
type pluginMap  map[string]ChannelPlugin

type channelPluginList []ChannelPlugin
type channelPluginMap  map[string]channelPluginList

type pluginManager struct {
	db         *DatabaseStruct
	bot        *Kabukibot
	dispatcher Dispatcher
	plugins    pluginList
	pluginMap  pluginMap
	loaded     channelPluginMap
}

func NewPluginManager(bot *Kabukibot, dispatcher Dispatcher, db *DatabaseStruct) *pluginManager {
	return &pluginManager{
		db,
		bot,
		dispatcher,
		make(pluginList, 0),
		make(pluginMap),
		make(channelPluginMap),
	}
}

func (self *pluginManager) Plugins() *pluginList {
	return &self.plugins
}

func (self *pluginManager) PluginMap() *pluginMap {
	return &self.pluginMap
}

func (self *pluginManager) PluginKeys() []string {
	list := make([]string, len(self.pluginMap))
	idx  := 0

	for key, _ := range self.pluginMap {
		list[idx] = key
		idx = idx + 1
	}

	return list
}

func (self *pluginManager) Plugin(key string) (c ChannelPlugin, ok bool) {
	c, ok = self.pluginMap[key]
	return
}

func (self *pluginManager) LoadedPlugins(channel *twitch.Channel) channelPluginList {
	list, ok := self.loaded[channel.Name]
	if !ok {
		list = make(channelPluginList, 0)
	}

	return list
}

func (self *pluginManager) LoadedPluginKeys(channel *twitch.Channel) []string {
	list   := self.LoadedPlugins(channel)
	result := make([]string, len(list))

	for idx, plugin := range list {
		result[idx] = plugin.Key()
	}

	return result
}

func (self *pluginManager) IsLoaded(channel *twitch.Channel, plugin string) bool {
	list, ok := self.loaded[channel.Name]
	if !ok {
		return false
	}

	for _, loaded := range list {
		if loaded.Key() == plugin {
			return true
		}
	}

	return false
}

func (self *pluginManager) AddPluginToChannel(channel *twitch.Channel, pluginKey string) bool {
	_, ok := self.pluginMap[pluginKey]
	if !ok {
		return false
	}

	if self.IsLoaded(channel, pluginKey) {
		return false
	}

	_, err := self.db.Exec("INSERT INTO plugin VALUES (channel, plugin)", channel.Name, pluginKey)
	if err != nil {
		log.Fatal("Could not add plugin to the database: " + err.Error())
	}

	self.loadPlugin(channel, pluginKey)

	return true
}

func (self *pluginManager) RemovePluginFromChannel(channel *twitch.Channel, pluginKey string) bool {
	_, ok := self.pluginMap[pluginKey]
	if !ok {
		return false
	}

	if !self.IsLoaded(channel, pluginKey) {
		return false
	}

	_, err := self.db.Exec("DELETE FROM plugin WHERE channel = ? AND plugin = ?", channel.Name, pluginKey)
	if err != nil {
		log.Fatal("Could not remove plugin from database: " + err.Error())
	}

	self.unloadPlugin(channel, pluginKey)

	return true
}

func (self *pluginManager) registerPlugin(plugin Plugin) {
	self.plugins = append(self.plugins, plugin)

	asserted, err := plugin.(ChannelPlugin)
	if !err {
		self.pluginMap[asserted.Key()] = asserted
	}
}

func (self *pluginManager) setup() {
	for _, plugin := range self.plugins {
		plugin.Setup(self.bot, self.dispatcher)
	}
}

func (self *pluginManager) setupChannel(channel *twitch.Channel) {
	// already loaded?
	_, ok := self.loaded[channel.Name]
	if ok {
		return
	}

	self.loaded[channel.Name] = make(channelPluginList, 0)

	rows, err := self.db.Query("SELECT plugin FROM plugin WHERE channel = ?", channel.Name)
	if err != nil {
		log.Fatal("Could not query the plugins: " + err.Error())
	}
	defer rows.Close()

	// collect the keys before loading the plugins, so we have this query
	// done before another (by plugins) starts
	keys := make([]string, 0)

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			log.Fatal(err)
		}

		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// now we can load the plugins
	for _, key := range keys {
		self.loadPlugin(channel, key)
	}
}

func (self *pluginManager) teardownChannel(channel *twitch.Channel) {
	for _, key := range self.LoadedPluginKeys(channel) {
		self.unloadPlugin(channel, key)
	}
}

func (self *pluginManager) loadPlugin(channel *twitch.Channel, key string) {
	// unknown plugin
	plugin, ok := self.pluginMap[key]
	if !ok {
		return
	}

	if self.IsLoaded(channel, key) {
		return
	}

	plugin.Load(channel, self.bot, self.dispatcher)
	self.loaded[channel.Name] = append(self.loaded[channel.Name], plugin)
}

func (self *pluginManager) unloadPlugin(channel *twitch.Channel, key string) {
	// unknown plugin
	plugin, ok := self.pluginMap[key]
	if !ok {
		return
	}

	// is this plugin loaded in this channel?
	loadedKeys := self.LoadedPluginKeys(channel)
	idx        := -1

	for i, ikey := range loadedKeys {
		if ikey == key {
			idx = i
			break
		}
	}

	if idx == -1 {
		return
	}

	plugin.Unload(channel, self.bot, self.dispatcher)

	if len(loadedKeys) > 1 {
		self.loaded[channel.Name] = append(self.loaded[channel.Name][:idx], self.loaded[channel.Name][(idx+1):]...)
	} else {
		delete(self.loaded, channel.Name)
	}
}
