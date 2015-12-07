package dictionary

import (
	"sort"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type pluginStruct struct {
	plugin.BasePlugin
	plugin.NilWorker

	dict *bot.Dictionary
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.dict = bot.Dictionary()
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}

func (self *pluginStruct) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
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

func (self *pluginStruct) handleSet(msg *bot.TextMessage, sender bot.Sender) {
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

func (self *pluginStruct) handleGet(msg *bot.TextMessage, sender bot.Sender) {
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

func (self *pluginStruct) handleKeys(msg *bot.TextMessage, sender bot.Sender) {
	keys := self.dict.Keys()

	sort.Strings(keys)

	sender.Respond("keys are: " + strings.Join(keys, ", "))
}
