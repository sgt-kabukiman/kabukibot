package sysinfo

import (
	"sync"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

// TODO: At some point, maybe don't use a mutex for the counter, but some
// channels and a counter goroutine

type pluginStruct struct {
	plugin.BasePlugin
	plugin.NilWorker

	bot      *bot.Kabukibot
	startup  time.Time
	messages int
	mutex    sync.Mutex
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.bot = bot
	self.startup = time.Now()
	self.mutex = sync.Mutex{}
	self.messages = 0
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}
