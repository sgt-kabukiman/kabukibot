package plugin

import "strings"
import "sort"
import "github.com/sgt-kabukiman/kabukibot/bot"

type DictionaryControlPlugin struct {
	bot    *bot.Kabukibot
	dict   *bot.Dictionary
	prefix string
}

func NewDictionaryControlPlugin() *DictionaryControlPlugin {
	return &DictionaryControlPlugin{}
}

func (self *DictionaryControlPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot    = bot
	self.dict   = bot.Dictionary()
	self.prefix = bot.Configuration().CommandPrefix

	d.OnCommand(self.onCommand, nil)
}

func (self *DictionaryControlPlugin) onCommand(cmd bot.Command) {
	if cmd.Processed() { return }

	// op-only
	if !self.bot.IsOperator(cmd.User().Name) {
		return
	}

	switch cmd.Command() {
		case self.prefix + "dict_set":  self.handleSet(cmd)
		case self.prefix + "dict_get":  self.handleGet(cmd)
		case self.prefix + "dict_keys": self.handleKeys(cmd)
	}
}

func (self *DictionaryControlPlugin) handleSet(cmd bot.Command) {
	args := cmd.Args()

	if len(args) < 2 {
		self.bot.Respond(cmd, "syntax is `!" + self.prefix + "dict_set key Your text here`.")
	}

	key    := args[0]
	value  := strings.Join(args[1:], " ")
	exists := self.dict.Has(key)

	self.dict.Set(key, value)

	if exists {
		self.bot.Respond(cmd, "replaced '" + key + "' with '" + value + "'.")
	} else {
		self.bot.Respond(cmd, "added '" + key + "' with '" + value + "'.")
	}
}

func (self *DictionaryControlPlugin) handleGet(cmd bot.Command) {
	args := cmd.Args()

	if len(args) < 1 {
		self.bot.Respond(cmd, "syntax is `!" + self.prefix + "dict_get key`; See !" + self.prefix + "dict_keys for a list of all keys.")
	}

	key := strings.ToLower(args[0])

	if self.dict.Has(key) {
		self.bot.Respond(cmd, key + " = " + self.dict.Get(key))
	} else {
		self.bot.Respond(cmd, "unknown key '" + key + "' given.")
	}
}

func (self *DictionaryControlPlugin) handleKeys(cmd bot.Command) {
	keys := self.dict.Keys()

	sort.Strings(keys)

	self.bot.Respond(cmd, "keys are: " + strings.Join(keys, ", "))
}
