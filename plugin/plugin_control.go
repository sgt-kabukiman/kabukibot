package plugin

import "strings"
import "sort"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type channelPluginStateMap map[string]bool

type PluginControlPlugin struct {
	bot    *bot.Kabukibot
	mngr   *bot.PluginManager
	acl    *bot.ACL
	prefix string
}

func NewPluginControlPlugin() *PluginControlPlugin {
	return &PluginControlPlugin{}
}

func (plugin *PluginControlPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	plugin.bot    = bot
	plugin.mngr   = bot.PluginManager()
	plugin.acl    = bot.ACL()
	plugin.prefix = bot.Configuration().CommandPrefix

	d.OnCommand(plugin.onCommand)
}

func (plugin *PluginControlPlugin) onCommand(cmd bot.Command) {
	c := cmd.Command()
	p := plugin.prefix

	// skip unwanted commands
	if c != p+"enable" && c != p+"disable" && c != p+"plugins" {
		return
	}

	// our commands are all priv only
	user := cmd.User()

	if !user.IsBroadcaster && !plugin.bot.IsOperator(user.Name) {
		return
	}

	channel := cmd.Channel()
	args    := cmd.Args()

	// send the list of available permissions
	if c == p+"plugins" {
		plugin.respondToListCommand(channel, args)
		return
	}

	// everything from now on requires at last a plugin key as the first parameter
	if len(args) == 0 {
		plugin.bot.Say(channel, "no plugin name given. See !" + p + "plugins for a list of available plugins.")
		return
	}

	// check the plugin
	pluginKey := strings.ToLower(args[0])
	allKeys   := plugin.getPluginKeys()
	found     := false

	for _, key := range allKeys {
		if key == pluginKey {
			found = true
			break
		}
	}

	if !found {
		plugin.bot.Say(channel, "invalid plugin (" + pluginKey + ") given.")
		return
	}

	mngr    := plugin.mngr
	message := ""

	// enable a plugin
	if c == p+"enable" {
		if mngr.AddPluginToChannel(channel, pluginKey) {
			message = "the plugin " + pluginKey + " has been enabled."
		} else {
			message = "the plugin " + pluginKey + " is already enabled in this channel."
		}
	} else { // disable a plugin
		if mngr.RemovePluginFromChannel(channel, pluginKey) {
			message = "the plugin " + pluginKey + " has been disabled."
		} else {
			message = "the plugin " + pluginKey + " is not enabled in this channel."
		}
	}

	plugin.bot.Say(channel, message)
}

func (plugin *PluginControlPlugin) respondToListCommand(channel *twitch.Channel, args []string) {
	plugins     := plugin.getPluginStates(channel)
	enabledOnly := len(args) > 0 && strings.ToLower(args[0]) == "enabled"
	nameList    := make([]string, 0)

	for pluginKey, enabled := range plugins {
		if enabledOnly {
			if enabled {
				nameList = append(nameList, pluginKey)
			}
		} else {
			if enabled {
				nameList = append(nameList, pluginKey + " (enabled)")
			} else {
				nameList = append(nameList, pluginKey + " (disabled)")
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
		plugin.bot.Say(channel, "There are no " + prefix + " plugins.")
	} else {
		plugin.bot.Say(channel, prefix + " plugins are: " + strings.Join(nameList, ", "))
	}
}

func (plugin *PluginControlPlugin) getPluginKeys() []string {
	keys := plugin.mngr.PluginKeys()

	sort.Strings(keys)

	return keys
}

func (plugin *PluginControlPlugin) getPluginStates(channel *twitch.Channel) channelPluginStateMap {
	result := make(channelPluginStateMap)

	for _, key := range plugin.getPluginKeys() {
		result[key] = plugin.mngr.IsLoaded(channel, key)
	}

	return result
}
