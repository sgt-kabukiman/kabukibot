package content

import (
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type worker struct {
	plugin.NilWorker

	acl    *bot.ACL
	dict   *bot.Dictionary
	plugin *pluginStruct
}

func (self *worker) Permissions() []string {
	return []string{self.plugin.permission}
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() {
		return
	}

	name := self.plugin.name

	if msg.IsFromOperator() {
		if msg.IsGlobalCommand(name + "_define") {
			msg.SetProcessed()

			args := msg.Arguments()
			if len(args) < 2 {
				sender.Respond("you must specify the new command name and the dictionary key it points to.")
				return
			}

			cmdName := args[0]
			dictKey := args[1]
			initial := strings.Join(args[2:], " ")

			if self.plugin.defineCommand(cmdName, dictKey, initial) {
				sender.Respond("new command !" + cmdName + " has been created.")
			} else {
				dictKey, _ := self.plugin.resolveCommand(cmdName)
				sender.Respond("!" + cmdName + " already exists and points to '" + dictKey + "'.")
			}

			return
		} else if msg.IsGlobalCommand(name + "_undefine") {
			msg.SetProcessed()

			args := msg.Arguments()
			if len(args) < 1 {
				sender.Respond("you must specify the command to remove.")
				return
			}

			cmdName := args[0]

			if self.plugin.undefineCommand(cmdName) {
				sender.Respond("the command !" + cmdName + " has been removed.")
			} else {
				sender.Respond("!" + cmdName + " does not exist or cannot be removed.")
			}

			return
		}
	}

	dictKey, exists := self.plugin.resolveCommand(msg.Command())
	if !exists {
		return
	}

	if !self.acl.IsAllowed(msg.User, self.plugin.permission) {
		return
	}

	response := self.dict.Get(dictKey)

	if len(response) > 0 {
		templater := bot.NewStringTemplater()

		sender.SendText(templater.Render(response))
	}

	msg.SetProcessed()
}
