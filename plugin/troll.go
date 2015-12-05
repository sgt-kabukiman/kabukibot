package plugin

import (
	"math/rand"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

var trollResponses = map[string][]string{
	"why": {
		// the very lazy man's RNG manipulation
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Because it's faster.",
		"Doing this manipulates the RNG in later parts of the game. Just wait a bit, you'll see.",
		"Doing this is literally just for the lulz of it.",
		"Doing this prevents viewers from noticing that the runner is actually using cheats.",
	},
	"system": {
		"This is on PC.",
		"This is on C64.",
		"This is on Magnavox Odyssey.",
		"This is on Atari 2600.",
		"This is on the Nintendo Entertainment System (NES).",
		"This is on Sega Genesis.",
		"This is on PlayStation 4.",
		"This is on ENIAC.",
		"This is on Zuse Z1.",
	},
	"song": {
		// "The current song is \"Sandstorm\" by Darude.",
		"The current song is \"Never Gonna Give You Up\" by Rick Astley.", // shoutouts to dfocus89
		"The current song is \"Inside Out\" by DotEXE.",                   // shoutouts to mhmd_fvc
		"The current song is \"Barbie Girl\" by Aqua.",                    // shoutouts to Eidgod
		"The current song is \"PON PON PON\" by Kyary Pamyu Pamyu.",
		"The current song is \"Bangarang\" by Skrillex.",
		"The current song is \"Hooked on a Feeling\" by David Hasselhoff.",
		"The current song is \"Judas\" by Lady Gaga.",
		"The current song is \"Friday\" by Rebecca Black.",
		"The current song is \"Dancing in the Street\" by David Bowie & Mick Jagger.",
		"Currently playing is Bach - Toccata and Fugue in D Minor.",
	},
}

type TrollPlugin struct{}

func NewTrollPlugin() *TrollPlugin {
	return &TrollPlugin{}
}

func (self *TrollPlugin) Name() string {
	return "troll"
}

func (self *TrollPlugin) Setup(bot *bot.Kabukibot) {
}

func (self *TrollPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &trollWorker{channel.ACL()}
}

type trollWorker struct {
	acl *bot.ACL
}

func (self *trollWorker) Enable() {
	// do nothing
}

func (self *trollWorker) Disable() {
	// do nothing
}

func (self *trollWorker) Part() {
	// do nothing
}

func (self *trollWorker) Shutdown() {
	// do nothing
}

func (self *trollWorker) Permissions() []string {
	return []string{"trolling"}
}

func (self *trollWorker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() || msg.IsFromBot() {
		return
	}

	cmd := msg.Command()
	if len(cmd) == 0 {
		return
	}

	responses, okay := trollResponses[cmd]
	if !okay {
		return
	}

	if self.acl.IsAllowed(msg.User, "trolling") {
		pos := rand.Intn(len(responses))
		sender.Respond(responses[pos])
	}

	msg.SetProcessed()
}
