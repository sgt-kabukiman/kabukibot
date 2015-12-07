package troll

import (
	"math/rand"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type worker struct {
	plugin.NilWorker

	acl *bot.ACL
}

func (self *worker) Permissions() []string {
	return []string{"trolling"}
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
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
