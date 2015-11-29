package plugin

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/srapi"
)

type speedruncomConfig struct {
	Interval int
	Mapping  map[string]map[string]string
}

type SpeedrunComPlugin struct {
	config speedruncomConfig
	dict   *bot.Dictionary
}

func NewSpeedrunComPlugin() *SpeedrunComPlugin {
	return &SpeedrunComPlugin{}
}

func (self *SpeedrunComPlugin) Name() string {
	return "speedruncom"
}

func (self *SpeedrunComPlugin) Permissions() []string {
	return []string{"use_speedruncom_commands"}
}

func (self *SpeedrunComPlugin) Setup(bot *bot.Kabukibot) {
	self.config = speedruncomConfig{}
	self.dict = bot.Dictionary()

	err := bot.Configuration().PluginConfig(self.Name(), &self.config)
	if err != nil {
		bot.Logger().Warn("Could not load 'speedruncom' plugin configuration: %s", err)
	}

	go self.updater()
}

func (self *SpeedrunComPlugin) updater() {
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

			for _, leaderboard := range leaderboards.Data {
				if len(leaderboard.Runs) == 0 {
					continue
				}

				cat, err := leaderboard.Category(srapi.NoEmbeds)
				if err != nil {
					continue
				}

				dictKey, okay := catList[cat.ID]
				if !okay {
					continue
				}

				wr := leaderboard.Runs[0]
				formatted := formatWorldRecord(&wr.Run, game, cat, nil, nil, nil)

				self.dict.Set(dictKey, formatted)
			}
		}

		time.Sleep(interval)
	}
}

func (self *SpeedrunComPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &speedruncomWorker{
		channel: channel.Name(),
		acl:     channel.ACL(),
	}
}

type speedruncomWorker struct {
	channel string
	acl     *bot.ACL
}

func (self *speedruncomWorker) Enable() {
	// nothing to do for us
}

func (self *speedruncomWorker) Disable() {
	// nothing to do for us
}

func (self *speedruncomWorker) Part() {
	// nothing to do for us
}

func (self *speedruncomWorker) Shutdown() {
	// nothing to do for us
}

func (self *speedruncomWorker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() || msg.IsFromBot() {
		return
	}

	if msg.IsCommand("wr") {
		self.handleWorldRecordCommand(msg, sender)
		msg.SetProcessed()
	}
}

var cleanerRegexp = regexp.MustCompile(`[^a-zA-Z0-9]`)

func (self *speedruncomWorker) handleWorldRecordCommand(msg *bot.TextMessage, sender bot.Sender) {
	if !self.acl.IsAllowed(msg.User, "use_speedrun_commands") {
		return
	}

	args := msg.Arguments()

	if len(args) == 0 {
		sender.Respond("you have to give me a game abbreviation.")
		return
	}

	gameIdentifier := args[0]

	var category *srapi.Category

	// try to find the game
	game, err := srapi.GameByAbbreviation(gameIdentifier, "categories")
	if err != nil {
		sender.Respond("I could not find a game with the abbreviation \"" + gameIdentifier + "\".")
		return
	}

	// assume all further args form the category, like "All Missions" or "Any%";
	// we normalise the value to make it -- hopefully -- easier to find the correct category

	if len(args) > 1 {
		catIdentifier := cleanerRegexp.ReplaceAllString(strings.ToLower(strings.Join(args[1:], "")), "")

		categories, err := game.Categories(nil, nil, srapi.NoEmbeds)
		catNames := []string{}

		if err == nil {
			for _, cat := range categories {
				id := cleanerRegexp.ReplaceAllString(strings.ToLower(cat.Name), "")

				if cat.Type == "per-game" {
					catNames = append(catNames, cat.Name)
				}

				if id == catIdentifier {
					category = cat
				}
			}
		}

		if category == nil {
			sender.Respond("I could not find a category named \"" + strings.Join(args[1:], " ") + "\". Available categories are: " + bot.HumanJoin(catNames, ", "))
			return
		} else if category.Type != "per-game" {
			sender.Respond(category.Name + " is a IL category; cannot report records for now. Sorry.")
			return
		}
	}

	var lb *srapi.Leaderboard

	// fetch the leaderboard, if possible (only available for games with full-game categories by default)
	if category == nil {
		lb, err = game.PrimaryLeaderboard(&srapi.LeaderboardOptions{Top: 1}, "players,platforms,regions,category")
		if err != nil || lb == nil {
			sender.Respond(game.Names.International + " does not use full-game categories by default, so I don't know what category or level you are referring to.")
			return
		}

		category, err = lb.Category(srapi.NoEmbeds)
		if err != nil {
			sender.Respond("the data from speedrun.com is invalid, cannot procede. Sorry. Try again later or a with different game.")
			return
		}
	} else {
		lb, err = category.PrimaryLeaderboard(&srapi.LeaderboardOptions{Top: 1}, "players,platforms,regions")
		if err != nil || lb == nil {
			sender.Respond(game.Names.International + " does not have runs for its \"" + category.Name + "\" category.")
			return
		}
	}

	// the leaderboard could be empty
	if len(lb.Runs) == 0 {
		sender.Respond(game.Names.International + ": " + category.Name + " does not have any matching runs yet.")
		return
	}

	// show only the first WR
	firstRun := lb.Runs[0].Run
	formatted := formatWorldRecord(&firstRun, game, category, nil, nil, nil)

	sender.SendText(formatted)
}

func formatWorldRecord(run *srapi.Run, game *srapi.Game, cat *srapi.Category, players []*srapi.Player, region *srapi.Region, platform *srapi.Platform) string {
	var err *srapi.Error

	if game == nil {
		game, err = run.Game(srapi.NoEmbeds)
		if err != nil {
			return "Could not fetch game."
		}
	}

	if cat == nil {
		cat, err = run.Category(srapi.NoEmbeds)
		if err != nil {
			return "Could not fetch category."
		}
	}

	formatted := fmt.Sprintf("WR for %s [%s] is %s", game.Names.International, cat.Name, run.Times.Primary.Format())

	if run.Times.IngameTime.Duration > 0 {
		formatted += " (" + run.Times.IngameTime.Format() + " IGT)"
	}

	// collect player names
	names := []string{}

	if len(players) == 0 {
		players, err = run.Players()
		if err != nil {
			return "Could not fetch players: " + err.Error()
		}
	}

	for _, player := range players {
		names = append(names, player.Name())
	}

	formatted += " by " + bot.HumanJoin(names, ", ")

	if run.Date != nil {
		now := time.Now()
		duration := int(now.Sub(run.Date.Time).Hours() / 24)
		date := ""

		switch duration {
		case 0:
			date = "today"
		case 1:
			date = "yesterday"
		case -1:
			date = "tomorrow"
		default:
			if duration > 0 {
				date = fmt.Sprintf("%d days ago", duration)
			} else {
				date = fmt.Sprintf("in %d days", -duration)
			}
		}

		formatted += ", " + date
	}

	// append platform info
	if platform == nil {
		platform, err = run.Platform()
		if err != nil {
			return "Could not fetch platform."
		}
	}

	showRegion := true

	if platform != nil {
		formatted = formatted + " (played on " + platform.Name
		showRegion = platform.ID != "8zjwp7vo" // do not show on PC
	}

	// append region info

	if showRegion {
		if region == nil {
			region, err = run.Region()
			if err != nil {
				return "Could not fetch region."
			}
		}

		if region != nil {
			formatted += ", " + region.Name
		}
	}

	if platform != nil {
		formatted += ")"
	}

	return formatted + "."
}
