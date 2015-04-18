package plugin

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/http"
import "net/url"
import "strconv"
import "strings"
import "time"
import "github.com/sgt-kabukiman/kabukibot/bot"

type srCategoryMap map[string]string
type srGameMap     map[string]srCategoryMap
type srSeriesMap   map[string]srGameMap

type srConfig struct {
	Interval int
	Mapping  srSeriesMap
}

type srRecord struct {
	Time    string
	TimeIGT string
	Date    string
	Player  string
}

type srGameLeaderboard   map[string]srRecord
type srSeriesLeaderboard map[string]srGameLeaderboard

type SpeedrunComPlugin struct {
	bot      *bot.Kabukibot
	log      bot.Logger
	dict     *bot.Dictionary
	interval time.Duration
	mapping  srSeriesMap
}

func NewSpeedrunComPlugin() *SpeedrunComPlugin {
	return &SpeedrunComPlugin{}
}

func (self *SpeedrunComPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot      = bot
	self.log      = bot.Logger()
	self.dict     = bot.Dictionary()
	self.interval = 15*time.Minute
	self.mapping  = make(srSeriesMap)

	data, exists := bot.Configuration().Plugins["speedruncom"]

	if exists {
		// very lazy hack because i could not figure out how to nicely type assert
		// the existing structure (which seems to be an endless map[string]interface{}
		// monster) to the srConfig struct
		config     := srConfig{}
		encoded, _ := json.Marshal(data)

		if json.Unmarshal(encoded, &config) != nil {
			bot.Logger().Fatal("Invalid speedrun.com configuration.")
		}

		self.interval = time.Duration(config.Interval) * time.Minute
		self.mapping  = config.Mapping
	}

	go self.updaterRoutine()
}

func (self *SpeedrunComPlugin) updaterRoutine() {
	for {
		for series, gameMap := range self.mapping {
			self.log.Info("Fetching leaderboard for '%s'...", series)

			// fetch URL
			response, err := http.Get("http://www.speedrun.com/api_records.php?series=" + url.QueryEscape(series))
			if err != nil {
				self.log.Error("Could not fetch leaderboard for '%s': %s", series, err.Error())
				continue
			}

			// read body
			body, err := ioutil.ReadAll(response.Body)
			response.Body.Close()

			if err != nil {
				self.log.Error("Could not read HTTP response body: %s", err.Error())
				continue
			}

			// parse leaderboard JSON
			leaderboard := srSeriesLeaderboard{}

			if json.Unmarshal(body, &leaderboard) != nil {
				self.log.Error("Received invalid JSON for leaderboard for '%s'.", series)
				continue
			}

			// update dictionary
			for game, catMap := range gameMap {
				categories, exists := leaderboard[game]
				if !exists {
					self.log.Warning("Game '%s' of '%s' was not found.", game, series)
					continue
				}

				for category, dictKey := range catMap {
					record, exists := categories[category]
					if !exists {
						self.log.Warning("Category '%s' in game '%s' of '%s' was not found.", category, game, series)
						continue
					}

					self.updateDictionary(dictKey, game, category, record)
				}
			}
		}

		<-time.After(self.interval)
	}
}

func (self *SpeedrunComPlugin) updateDictionary(key string, game string, category string, record srRecord) {
	runtime, err := strconv.ParseFloat(record.Time, 32)
	if err != nil {
		self.log.Warning("No valid time found for %s: %s", key, err.Error())
		return
	}

	date, err := strconv.Atoi(record.Date)
	if err != nil {
		self.log.Warning("No valid time found for %s: %s", key, err.Error())
		return
	}

	// it's okay if there is no [valid] igt runtime
	runtimeIGT, err := strconv.ParseFloat(record.TimeIGT, 32)
	if err != nil {
		runtimeIGT = 0
	}

	// because I like it better this way
	if category == "Any%" {
		category = "any%"
	}

	game  = strings.Replace(game, "Grand Theft Auto", "GTA", -1)
	text := fmt.Sprintf("WR for %s %s is %s", game, category, bot.SecondsToRunTime(float32(runtime)))

	if runtimeIGT > 0 {
		text += fmt.Sprintf(" (%s IGT)", bot.SecondsToRunTime(float32(runtimeIGT)))
	}

	text += fmt.Sprintf(" by %s, <reldate>%s</reldate>", record.Player, time.Unix(int64(date), 0).Format("2 Jan. 2006"))

	// self.dict.Set(key, text)
	fmt.Printf("%s = %s\n", key, text)
}
