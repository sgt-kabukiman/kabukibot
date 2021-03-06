package blacklist

import (
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type pluginStruct struct {
	plugin.BasePlugin
	plugin.NilWorker

	db    *sqlx.DB
	log   bot.Logger
	users []string
	bot   string
	mutex sync.RWMutex
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.db = bot.Database()
	self.log = bot.Logger()
	self.bot = strings.ToLower(bot.BotUsername())
	self.mutex = sync.RWMutex{}

	self.loadBlacklist()
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return self
}
