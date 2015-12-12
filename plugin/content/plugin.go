package content

import (
	"strings"
	"sync"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin/speedruncom"
)

type command struct {
	dictKey string
	fixed   bool
}

type pluginStruct struct {
	name       string
	permission string
	srcom      bool
	srPrefix   string
	dict       *bot.Dictionary
	commands   map[string]command
	cmdMutex   sync.RWMutex
}

func newPlugin(name string, perm string, srcom bool, srprefix string) *pluginStruct {
	return &pluginStruct{
		name:       name,
		permission: perm,
		srcom:      srcom,
		srPrefix:   srprefix,
		commands:   make(map[string]command),
	}
}

func NewGTAPlugin() *pluginStruct {
	return newPlugin("gta", "gta_commands", true, "gta_")
}

func NewCrashPlugin() *pluginStruct {
	return newPlugin("crash", "crash_commands", true, "crash_")
}

func NewChattyPlugin() *pluginStruct {
	return newPlugin("chatty", "chatty_commands", false, "")
}

func NewESAPlugin() *pluginStruct {
	return newPlugin("esa", "esa_commands", false, "")
}

func NewSDAPlugin() *pluginStruct {
	return newPlugin("sda", "sda_commands", false, "")
}

func NewEGGSPlugin() *pluginStruct {
	return newPlugin("eggs", "eggs_commands", false, "")
}

func (self *pluginStruct) Name() string {
	return self.name
}

func (self *pluginStruct) cmdKeyPrefix() string {
	return self.name + "_cmd_"
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.dict = bot.Dictionary()

	self.cmdMutex.Lock()
	defer self.cmdMutex.Unlock()

	prefix := self.cmdKeyPrefix()

	// load dictionary entries for the FAQ commands
	for _, key := range self.dict.Keys() {
		if strings.HasPrefix(key, prefix) {
			cmdName := strings.TrimPrefix(key, prefix)
			targetKey := self.dict.Get(key)

			self.commands[cmdName] = command{
				dictKey: targetKey,
				fixed:   false,
			}
		}
	}

	if self.srcom {
		var srPlugin *speedruncom.Plugin

		for _, w := range bot.Plugins() {
			asserted, okay := w.(*speedruncom.Plugin)
			if okay {
				srPlugin = asserted
				break
			}
		}

		if srPlugin == nil {
			panic("Cannot run the " + self.name + " plugin without the speedrun.com plugin.")
		}

		// collect WR commands defined in the bot config file
		for cmd, dictKey := range srPlugin.CollectCommands(self.srPrefix) {
			self.commands[cmd] = command{
				dictKey: dictKey,
				fixed:   true,
			}
		}
	}
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &worker{
		acl:    channel.ACL(),
		dict:   self.dict,
		plugin: self,
	}
}

func (self *pluginStruct) resolveCommand(cmd string) (string, bool) {
	self.cmdMutex.RLock()
	defer self.cmdMutex.RUnlock()

	c, okay := self.commands[cmd]

	return c.dictKey, okay
}

func (self *pluginStruct) defineCommand(cmd string, dictKey string, initialValue string) bool {
	self.cmdMutex.Lock()
	defer self.cmdMutex.Unlock()

	_, exists := self.commands[cmd]
	if exists {
		return false
	}

	self.commands[cmd] = command{
		dictKey: dictKey,
		fixed:   false,
	}

	self.dict.Set(self.cmdKeyPrefix()+cmd, dictKey)

	if len(initialValue) > 0 {
		// do not overwrite existing values
		existing := self.dict.Get(dictKey)

		if len(existing) == 0 {
			self.dict.Set(dictKey, initialValue)
		}
	}

	return true
}

func (self *pluginStruct) undefineCommand(cmd string) bool {
	self.cmdMutex.Lock()
	defer self.cmdMutex.Unlock()

	c, exists := self.commands[cmd]
	if !exists || c.fixed {
		return false
	}

	delete(self.commands, cmd)

	self.dict.Delete(self.cmdKeyPrefix() + cmd)

	// do not remove the actual FAQ content, in case multiple commands may point to it;
	// plus it does not relly hurt to have unused dict keys lying around.
	// self.dict.Delete(targetKey)

	return true
}
