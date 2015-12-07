package custom_commands

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	"github.com/sgt-kabukiman/kabukibot/plugin/acl"
)

type worker struct {
	plugin.NilWorker

	channel   bot.Channel
	acl       *bot.ACL
	aclWorker *acl.Worker
	db        *sqlx.DB
	commands  map[string]string
}

type ccDbStruct struct {
	Command string
	Message string
}

func (self *worker) Enable() {
	list := make([]ccDbStruct, 0)
	self.db.Select(&list, "SELECT command, message FROM custom_commands WHERE channel = ? ORDER BY command", self.channel.Name())

	self.commands = make(map[string]string)

	for _, item := range list {
		self.commands[item.Command] = item.Message
	}

	for _, w := range self.channel.Workers() {
		asserted, okay := w.(*acl.Worker)
		if okay {
			self.aclWorker = asserted
			break
		}
	}

	if self.aclWorker == nil {
		panic("Cannot run the custom commands plugin without the ACL plugin.")
	}
}

func (self *worker) Permissions() []string {
	permissions := []string{"configure_custom_commands", "configure_custom_commands_acl", "list_custom_commands"}

	for cmd := range self.commands {
		permissions = append(permissions, permissionForCommand(cmd))
	}

	return permissions
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() || msg.IsFromBot() {
		return
	}

	command := msg.Command()
	if len(command) == 0 {
		return
	}

	isSysCmd := isPluginCommand(command)
	response, isUserCmd := self.commands[command]

	if !isSysCmd && !isUserCmd {
		return
	}

	msg.SetProcessed()

	if !self.acl.IsAllowed(msg.User, permissionForCommand(command)) {
		return
	}

	switch command {
	case "cc_list":
		self.respondList(sender)

	case "cc_allow":
		fallthrough
	case "cc_deny":
		fallthrough
	case "cc_get":
		fallthrough
	case "cc_set":
		fallthrough
	case "cc_del":
		args := msg.Arguments()
		if len(args) < 1 {
			sender.Respond("no command name given.")
			return
		}

		cc := normalizeCommand(args[0])
		if len(cc) < 1 {
			sender.Respond("invalid command name given.")
			return
		}

		switch command {
		case "cc_allow":
			self.respondAllowDeny("allow", cc, args[1:], sender)
		case "cc_deny":
			self.respondAllowDeny("deny", cc, args[1:], sender)
		case "cc_get":
			self.respondGet(cc, sender)
		case "cc_set":
			self.respondSet(cc, args[1:], sender)
		case "cc_del":
			self.respondDelete(cc, sender)
		}

	default:
		sender.SendText(response)
	}
}

func (self *worker) respondList(sender bot.Sender) {
	var commands []string

	for cmd, _ := range self.commands {
		commands = append(commands, "!"+cmd)
	}

	if len(commands) == 0 {
		sender.Respond("no custom commands have been defined yet.")
	} else {
		sender.Respond(fmt.Sprintf("this channel's custom commands are: %s", bot.HumanJoin(commands, ", ")))
	}
}

func (self *worker) respondAllowDeny(kind string, cmd string, args []string, sender bot.Sender) {
	_, exists := self.commands[cmd]
	if !exists {
		sender.Respond("there is no custom command named '" + cmd + "'.")
		return
	}

	permission := permissionForCommand(cmd)

	self.aclWorker.HandleAllowDeny(kind == "allow", permission, args, sender, "!"+cmd)
}

func (self *worker) respondGet(cmd string, sender bot.Sender) {
	response, exists := self.commands[cmd]
	if !exists {
		sender.Respond("there is no custom command named '" + cmd + "'.")
		return
	}

	sender.Respond("!" + cmd + " = " + response)
}

func (self *worker) respondSet(cmd string, args []string, sender bot.Sender) {
	if len(args) < 1 {
		sender.Respond("you did not give any response text for the new !" + cmd + " command.")
		return
	}

	if isPluginCommand(cmd) {
		sender.Respond("you cannot overwrite cc_* commands.")
		return
	}

	_, exists := self.commands[cmd]
	response := strings.Join(args, " ")

	self.commands[cmd] = response

	if exists {
		sender.Respond("command !" + cmd + " has been updated.")

		_, err := self.db.Exec("UPDATE custom_commands SET message = ? WHERE channel = ? AND command = ?", response, self.channel.Name(), cmd)
		if err != nil {
			log.Fatal("Could not update new custom command: " + err.Error())
		}
	} else {
		sender.Respond("command !" + cmd + " has been created. Do not forget to set permissions via `!cc_allow " + cmd + " $mods,someone,etc`.")

		_, err := self.db.Exec("INSERT INTO custom_commands (channel, command, message) VALUES (?, ?, ?)", self.channel.Name(), cmd, response)
		if err != nil {
			log.Fatal("Could not store new custom command: " + err.Error())
		}
	}
}

func (self *worker) respondDelete(cmd string, sender bot.Sender) {
	_, exists := self.commands[cmd]
	if !exists {
		sender.Respond("there is no custom command named '" + cmd + "'.")
		return
	}

	sender.Respond("!" + cmd + " has been deleted.")

	delete(self.commands, cmd)

	// cleanup database
	_, err := self.db.Exec("DELETE FROM custom_commands WHERE channel = ? AND command = ?", self.channel.Name(), cmd)
	if err != nil {
		log.Fatal("Could not delete new custom command: " + err.Error())
	}

	// cleanup ACL entries
	self.acl.DeletePermission(permissionForCommand(cmd))
}

func isPluginCommand(cmd string) bool {
	return cmd == "cc_set" || cmd == "cc_get" || cmd == "cc_del" || cmd == "cc_list" || cmd == "cc_allow" || cmd == "cc_deny"
}

func requiredPermission(cmd string) string {
	if cmd == "cc_allow" || cmd == "cc_deny" {
		return "configure_custom_commands_acl"
	} else if cmd == "cc_list" {
		return "list_custom_commands"
	} else if isPluginCommand(cmd) {
		return "configure_custom_commands"
	} else {
		return permissionForCommand(cmd)
	}
}

func permissionForCommand(cmd string) string {
	return "use_" + cmd + "_cmd"
}

var ccCommandCleaner = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func normalizeCommand(cmd string) string {
	return strings.ToLower(ccCommandCleaner.ReplaceAllString(cmd, ""))
}
