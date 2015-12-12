package speedruncom

import (
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/srapi"
)

type categoryConfig struct {
	DictKey  string `yaml:"dict"`
	Commands []string
}

type gameConfig map[string]categoryConfig

type speedruncomConfig struct {
	Interval int
	Mapping  map[string]gameConfig
}

type Plugin struct {
	config speedruncomConfig
	dict   *bot.Dictionary
}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (self *Plugin) Name() string {
	return "speedruncom"
}

func (self *Plugin) Setup(bot *bot.Kabukibot) {
	self.config = speedruncomConfig{}
	self.dict = bot.Dictionary()

	err := bot.Configuration().PluginConfig(self.Name(), &self.config)
	if err != nil {
		bot.Logger().Warning("Could not load 'speedruncom' plugin configuration: %s", err)
	}

	go self.updater()
}

func (self *Plugin) updater() {
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

				catConfig, okay := catList[cat.ID]
				if !okay {
					return true
				}

				wr := lb.Runs[0]
				formatted := formatWorldRecord(&wr.Run, game, cat, nil, nil, nil)

				self.dict.Set(catConfig.DictKey, formatted)

				return true
			})

			time.Sleep(5 * time.Second)
		}

		time.Sleep(interval)
	}
}

func (self *Plugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &worker{
		channel: channel.Name(),
		acl:     channel.ACL(),
	}
}

func (self *Plugin) CollectCommands(dictKeyPrefix string) map[string]string {
	result := make(map[string]string)

	for _, catList := range self.config.Mapping {
		for _, catConfig := range catList {
			if len(dictKeyPrefix) == 0 || strings.HasPrefix(catConfig.DictKey, dictKeyPrefix) {
				for _, cmd := range catConfig.Commands {
					result[cmd] = catConfig.DictKey
				}
			}
		}
	}

	return result
}
