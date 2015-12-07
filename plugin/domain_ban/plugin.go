package domain_ban

import (
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
	return "domain_ban"
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.db = bot.Database()
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &worker{
		channel: channel.Name(),
		acl:     channel.ACL(),
		db:      self.db,
	}
}
