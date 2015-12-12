package speedruncom

import (
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/srapi"
)

type speedruncomConfig struct {
	Interval int
	Mapping  map[string]map[string]string
}

type pluginStruct struct {
	config speedruncomConfig
	dict   *bot.Dictionary
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{}
}

func (self *pluginStruct) Name() string {
	return "speedruncom"
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.config = speedruncomConfig{}
	self.dict = bot.Dictionary()

	err := bot.Configuration().PluginConfig(self.Name(), &self.config)
	if err != nil {
		bot.Logger().Warning("Could not load 'speedruncom' plugin configuration: %s", err)
	}

	go self.updater()
}

func (self *pluginStruct) updater() {
	interval := time.Duration(self.config.Interval * int(time.Minute))

	for {
		for gameID, catList := range self.config.Mapping {
			game, err := srapi.GameByID(gameID, srapi.NoEmbeds)
			if err != nil {
				continue
			}

			leaderboards, err := game.Records(nil, "players,regions,platforms,category")
			if err != nil {
				continue
			}

			leaderboards.Walk(func(lb *srapi.Leaderboard) bool {
				if len(lb.Runs) == 0 {
					return true
				}

				cat, err := lb.Category(srapi.NoEmbeds)
				if err != nil {
					return true
				}

				dictKey, okay := catList[cat.ID]
				if !okay {
					return true
				}

				wr := lb.Runs[0]
				formatted := formatWorldRecord(&wr.Run, game, cat, nil, nil, nil)

				self.dict.Set(dictKey, formatted)

				return true
			})

			time.Sleep(5 * time.Second)
		}

		time.Sleep(interval)
	}
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &worker{
		channel: channel.Name(),
		acl:     channel.ACL(),
	}
}
