package plugin

import (
	"sort"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type PluginControlPlugin struct {
	bot     *bot.Kabukibot
	prefix  string
	plugins []bot.Plugin
}

func NewPluginControlPlugin() *PluginControlPlugin {
	return &PluginControlPlugin{}
}

func (self *PluginControlPlugin) Name() string {
	return ""
}

func (self *PluginControlPlugin) Permissions() []string {
	return []string{}
}

func (self *PluginControlPlugin) Setup(bot *bot.Kabukibot) {
	self.bot = bot
	self.prefix = bot.Configuration().CommandPrefix
	self.plugins = bot.Plugins()
}

func (self *PluginControlPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &pluginControlWorker{
		bot:     self.bot,
		prefix:  self.prefix,
		channel: channel,
		plugins: self.plugins,
	}
}

type pluginControlWorker struct {
	bot     *bot.Kabukibot
	prefix  string
	channel bot.Channel
	plugins []bot.Plugin
}

func (self *pluginControlWorker) Enable() {
	// nothing to do for us
}

func (self *pluginControlWorker) Disable() {
	// nothing to do for us
}

func (self *pluginControlWorker) Part() {
	// nothing to do for us
}

func (self *pluginControlWorker) Shutdown() {
	// nothing to do for us
}

func (self *pluginControlWorker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() {
		return
	}

	// skip unwanted commands
	if !msg.IsGlobalCommand("enable") && !msg.IsGlobalCommand("disable") && !msg.IsGlobalCommand("plugins") {
		return
	}

	// our commands are all priv-only
	if !msg.IsFromBroadcaster() && !msg.IsFromOperator() {
		return
	}

	args := msg.Arguments()

	// send the list of available permissions
	if msg.IsGlobalCommand("plugins") {
		self.respondToListCommand(args, sender)
		return
	}

	// everything from now on requires at last a plugin key as the first parameter
	if len(args) == 0 {
		sender.Respond("no plugin name given. See !" + self.prefix + "plugins for a list of available plugins.")
		return
	}

	// check the plugin
	pluginKey := strings.ToLower(args[0])
	allKeys := self.pluginKeys()
	found := false

	for _, key := range allKeys {
		if key == pluginKey {
			found = true
			break
		}
	}

	if !found {
		sender.Respond("invalid plugin \"" + pluginKey + "\" given.")
		return
	}

	message := ""

	// enable a plugin
	if msg.IsGlobalCommand("enable") {
		if self.channel.EnablePlugin(pluginKey) {
			message = "the plugin " + pluginKey + " has been enabled."
		} else {
			message = "the plugin " + pluginKey + " is already enabled in this channel."
		}
	} else { // disable a plugin
		if self.channel.DisablePlugin(pluginKey) {
			message = "the plugin " + pluginKey + " has been disabled."
		} else {
			message = "the plugin " + pluginKey + " is not enabled in this channel."
		}
	}

	sender.Respond(message)
}

func (self *pluginControlWorker) respondToListCommand(args []string, sender bot.Sender) {
	plugins := self.pluginStates()
	enabledOnly := len(args) > 0 && strings.ToLower(args[0]) == "enabled"
	nameList := make([]string, 0)

	for pluginKey, enabled := range plugins {
		if enabledOnly {
			if enabled {
				nameList = append(nameList, pluginKey)
			}
		} else {
			if enabled {
				nameList = append(nameList, pluginKey+" (enabled)")
			} else {
				nameList = append(nameList, pluginKey+" (disabled)")
			}
		}
	}

	var prefix string
	if enabledOnly {
		prefix = "enabled"
	} else {
		prefix = "available"
	}

	if len(nameList) == 0 {
		sender.Respond("there are no " + prefix + " plugins.")
	} else {
		sender.Respond(prefix + " plugins are: " + strings.Join(nameList, ", "))
	}
}

func (self *pluginControlWorker) pluginKeys() []string {
	keys := make([]string, 0)

	for _, plugin := range self.plugins {
		if len(plugin.Name()) > 0 {
			keys = append(keys, plugin.Name())
		}
	}

	sort.Strings(keys)

	return keys
}

func (self *pluginControlWorker) pluginStates() map[string]bool {
	result := make(map[string]bool)

	for _, plugin := range self.plugins {
		if len(plugin.Name()) > 0 {
			result[plugin.Name()] = false
		}
	}

	for _, plugin := range self.channel.Plugins() {
		if len(plugin.Name()) > 0 {
			result[plugin.Name()] = true
		}
	}

	return result
}
