package speedruncom

import (
	"fmt"
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/srapi"
)

func formatWorldRecord(leaderboard *srapi.Leaderboard, runIdx int) string {
	if runIdx >= len(leaderboard.Runs) {
		return fmt.Sprintf("There is no %d. run in this leaderboard.", runIdx+1)
	}

	ranked := leaderboard.Runs[runIdx]
	run := ranked.Run

	// get the game, we always need that
	game, err := leaderboard.Game(srapi.NoEmbeds)
	if err != nil {
		return "Could not fetch game: " + err.Error()
	}

	// same goes for the category
	cat, err := leaderboard.Category(srapi.NoEmbeds)
	if err != nil {
		return "Could not fetch category: " + err.Error()
	}

	// build the beginning of the formatted string
	primary := run.Times.Primary.Format()
	formatted := fmt.Sprintf("WR for %s [%s] is %s", game.Names.International, cat.Name, primary)

	if run.Times.IngameTime.Duration > 0 {
		igt := run.Times.IngameTime.Format()

		if igt != primary {
			formatted += " (" + igt + " IGT)"
		}
	}

	// append players; take embedded players from the leaderboard instead of
	// re-embedding them in the run, which would cause another request
	allPlayers := leaderboard.Players()

	// collect players that took part in this run
	var participants []string

	playerLinks, err := run.PlayerLinks()
	if err != nil {
		return "Could not fetch player links: " + err.Error()
	}

	allPlayers.Walk(func(p *srapi.Player) bool {
		for _, link := range playerLinks {
			if p.User != nil && p.User.ID == link.ID {
				participants = append(participants, p.Name())
			} else if p.Guest != nil && p.Guest.Name == link.Name {
				participants = append(participants, p.Name())
			}
		}

		return true
	})

	formatted += " by " + bot.HumanJoin(participants, ", ")

	// append the relative run date, e.g. "100 days ago"
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
	platformID := run.System.Platform
	showRegion := true

	var sysInfo []string

	if len(platformID) > 0 {
		for _, platform := range leaderboard.Platforms().Data {
			if platform.ID == platformID {
				sysInfo = append(sysInfo, "played on "+platform.Name)
				showRegion = platform.ID != "8zjwp7vo" // do not show on PC
				break
			}
		}
	}

	// append region info
	if showRegion {
		regionID := run.System.Region

		if len(regionID) > 0 {
			for _, region := range leaderboard.Regions().Data {
				if region.ID == regionID {
					sysInfo = append(sysInfo, region.Name)
					break
				}
			}
		}
	}

	if len(sysInfo) > 0 {
		formatted += " (" + strings.Join(sysInfo, ", ") + ")"
	}

	return formatted + "."
}
