package plugin

import (
	"sort"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type DictionaryPlugin struct {
	dict     *bot.Dictionary
	operator string
}

func NewDictionaryPlugin() *DictionaryPlugin {
	return &DictionaryPlugin{}
}

func (self *DictionaryPlugin) Name() string {
	return ""
}

func (self *DictionaryPlugin) Permissions() []string {
	return []string{}
}

func (self *DictionaryPlugin) Setup(bot *bot.Kabukibot) {
	self.dict = bot.Dictionary()
	self.operator = bot.OpUsername()
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

func (self *DictionaryPlugin) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	// op-only
	if !msg.IsFrom(self.operator) {
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
		sender.SendText("no text given.")
		return
	}

	key := args[0]
	value := strings.Join(args[1:], " ")
	exists := self.dict.Has(key)

	self.dict.Set(key, value)

	if exists {
		sender.SendText("replaced '" + key + "' with '" + value + "'.")
	} else {
		sender.SendText("added '" + key + "' with '" + value + "'.")
	}
}

func (self *DictionaryPlugin) handleGet(msg *bot.TextMessage, sender bot.Sender) {
	args := msg.Arguments()

	if len(args) < 1 {
		sender.SendText("no key given.")
		return
	}

	key := strings.ToLower(args[0])

	if self.dict.Has(key) {
		sender.SendText(key + " = " + self.dict.Get(key))
	} else {
		sender.SendText("unknown key '" + key + "' given.")
	}
}

func (self *DictionaryPlugin) handleKeys(msg *bot.TextMessage, sender bot.Sender) {
	keys := self.dict.Keys()

	sort.Strings(keys)

	sender.SendText("keys are: " + strings.Join(keys, ", "))
}
