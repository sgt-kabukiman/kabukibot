package plugin

import (
	"sort"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type DictionaryPlugin struct {
	dict *bot.Dictionary
}

func NewDictionaryPlugin() *DictionaryPlugin {
	return &DictionaryPlugin{}
}

func (self *DictionaryPlugin) Name() string {
	return ""
}

func (self *DictionaryPlugin) Setup(bot *bot.Kabukibot) {
	self.dict = bot.Dictionary()
}

func (self *DictionaryPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *DictionaryPlugin) Enable() {
	// nothing to do for us
}

func (self *DictionaryPlugin) Disable() {
	// nothing to do for us
}

func (self *DictionaryPlugin) Part() {
	// nothing to do for us
}

func (self *DictionaryPlugin) Shutdown() {
	// nothing to do for us
}

func (self *DictionaryPlugin) Permissions() []string {
	return []string{}
}

func (self *DictionaryPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	// op-only
	if !msg.IsFromOperator() {
		return
	}

	if msg.IsGlobalCommand("dict_set") {
		self.handleSet(msg, sender)
	} else if msg.IsGlobalCommand("dict_get") {
		self.handleGet(msg, sender)
	} else if msg.IsGlobalCommand("dict_keys") {
		self.handleKeys(msg, sender)
	}
}

func (self *DictionaryPlugin) handleSet(msg *bot.TextMessage, sender bot.Sender) {
	args := msg.Arguments()

	if len(args) < 2 {
		sender.Respond("you have not given any text.")
		return
	}

	key := args[0]
	value := strings.Join(args[1:], " ")
	exists := self.dict.Has(key)

	self.dict.Set(key, value)

	if exists {
		sender.Respond("replaced '" + key + "' with '" + value + "'.")
	} else {
		sender.Respond("added '" + key + "' with '" + value + "'.")
	}
}

func (self *DictionaryPlugin) handleGet(msg *bot.TextMessage, sender bot.Sender) {
	args := msg.Arguments()

	if len(args) < 1 {
		sender.Respond("you have not given a key.")
		return
	}

	key := strings.ToLower(args[0])

	if self.dict.Has(key) {
		sender.Respond(key + " = " + self.dict.Get(key))
	} else {
		sender.Respond("the key '" + key + "' does not exist.")
	}
}

func (self *DictionaryPlugin) handleKeys(msg *bot.TextMessage, sender bot.Sender) {
	keys := self.dict.Keys()

	sort.Strings(keys)

	sender.Respond("keys are: " + strings.Join(keys, ", "))
}
