package emote_counter

import (
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
)

type pluginStruct struct {
	db *sqlx.DB
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Name() string {
	return "emote_counter"
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.db = bot.Database()
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &worker{
		channel:     channel.Name(),
		acl:         channel.ACL(),
		db:          self.db,
		syncing:     nil,
		stopSyncing: nil,
		queue:       make(chan *bot.TextMessage, 50),
		mutex:       sync.RWMutex{},
	}
}
