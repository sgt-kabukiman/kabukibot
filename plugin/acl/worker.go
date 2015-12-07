package acl

import (
	"regexp"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type Worker struct {
	plugin.NilWorker

	bot     *bot.Kabukibot
	channel bot.Channel
}

var permRegex = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func (self *Worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() {
		return
	}

	// skip unwanted commands
	if !msg.IsGlobalCommand("allow") && !msg.IsGlobalCommand("deny") && !msg.IsGlobalCommand("permissions") && !msg.IsGlobalCommand("allowed") {
		return
	}

	// our commands are all priv-only
	if !msg.IsFromBroadcaster() && !msg.IsFromOperator() {
		return
	}

	// send the list of available permissions
	if msg.IsGlobalCommand("permissions") {
		permissions := self.collectPermissions()

		if len(permissions) == 0 {
			sender.Respond("there are no permissions available to be configured.")
		} else {
			sender.Respond("available permissions are: " + strings.Join(permissions, ", "))
		}

		return
	}

	// everything from now on requires at last a permission as the first parameter
	args := msg.Arguments()
	if len(args) == 0 {
		sender.Respond("no permission name given.")
		return
	}

	// check the permission
	permission := strings.ToLower(permRegex.ReplaceAllString(args[0], ""))
	permissions := self.collectPermissions()

	if len(permission) == 0 {
		sender.Respond("invalid (no) permission given.")
		return
	}

	found := false

	for _, p := range permissions {
		if p == permission {
			found = true
			break
		}
	}

	if !found {
		sender.Respond("invalid permission (" + permission + ") given.")
		return
	}

	// send the list of usernames and groups that have been granted the X permission
	acl := self.channel.ACL()

	if msg.IsGlobalCommand("allowed") {
		users := acl.AllowedUsers(permission)

		if len(permissions) == 0 {
			sender.Respond("\"" + permission + "\" is granted to nobody at the moment, only you can use it.")
		} else {
			sender.Respond("\"" + permission + "\" is granted to " + strings.Join(users, ", "))
		}

		return
	}

	// no user ident(s) given
	if len(args) == 1 {
		sender.Respond("no groups/usernames given. Group names are " + strings.Join(bot.ACLGroups(), ", ") + ".")
		return
	}

	self.HandleAllowDeny(msg.IsGlobalCommand("allow"), permission, args[1:], sender, permission)
}

var userIdentRegex = regexp.MustCompile(`[^a-zA-Z0-9_$,]`)
var userNameRegex = regexp.MustCompile(`[^a-z0-9_]`)

// HandleAllowDeny is exported because the custom commands plugin re-uses it. #cheating
func (self *Worker) HandleAllowDeny(allow bool, permission string, args []string, sender bot.Sender, permisionName string) {
	// normalize the arguments into a single array of (possibly bogus) idents
	args = strings.Split(strings.ToLower(userIdentRegex.ReplaceAllString(strings.Join(args, ","), "")), ",")

	if len(args) == 0 {
		sender.Respond("invalid groups/usernames. Use a comma separated list if you give multiple.")
		return
	}

	processed := make([]string, 0)
	acl := self.channel.ACL()

	for i, ident := range args {
		if acl.IsUsername(ident) {
			ident = userNameRegex.ReplaceAllString(ident, "")

			// if we removed bogus characters, discard the user ident to not accidentally grant or revoke permissions
			if ident != args[i] {
				continue
			}
		}

		if allow {
			if acl.Allow(ident, permission) {
				processed = append(processed, ident)
			}
		} else {
			if acl.Deny(ident, permission) {
				processed = append(processed, ident)
			}
		}
	}

	if len(processed) == 0 {
		sender.Respond("no changes needed.")
	} else if allow {
		sender.Respond("granted permission for " + permisionName + " to " + bot.HumanJoin(processed, ", ") + ".")
	} else {
		sender.Respond("revoked permission for " + permisionName + " from " + bot.HumanJoin(processed, ", ") + ".")
	}
}

func (self *Worker) collectPermissions() []string {
	result := make([]string, 0)

	for _, worker := range self.channel.Workers() {
		for _, perm := range worker.Permissions() {
			result = append(result, perm)
		}
	}

	return result
}
