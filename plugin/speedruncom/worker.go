package speedruncom

import (
	"regexp"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	"github.com/sgt-kabukiman/srapi"
)

type worker struct {
	plugin.NilWorker

	channel string
	acl     *bot.ACL
}

func (self *worker) Permissions() []string {
	return []string{"use_speedruncom_commands"}
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() || msg.IsFromBot() {
		return
	}

	if msg.IsCommand("wr") {
		self.handleWorldRecordCommand(msg, sender)
		msg.SetProcessed()
	}
}

var cleanerRegexp = regexp.MustCompile(`[^a-zA-Z0-9]`)

func (self *worker) handleWorldRecordCommand(msg *bot.TextMessage, sender bot.Sender) {
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
			categories.Walk(func(cat *srapi.Category) bool {
				id := cleanerRegexp.ReplaceAllString(strings.ToLower(cat.Name), "")

				if cat.Type == "per-game" {
					catNames = append(catNames, cat.Name)
				}

				if id == catIdentifier {
					category = cat
				}

				return true
			})
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
		lb, err = game.PrimaryLeaderboard(&srapi.LeaderboardOptions{Top: 1}, "players,regions,platforms,category,game")
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
		lb, err = category.PrimaryLeaderboard(&srapi.LeaderboardOptions{Top: 1}, "players,regions,platforms,category,game")
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
	formatted := formatWorldRecord(lb, 0)

	sender.SendText(formatted)
}
