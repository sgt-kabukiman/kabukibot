package plugin

import (
	"regexp"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type ACLPlugin struct {
	bot *bot.Kabukibot
}

func NewACLPlugin() *ACLPlugin {
	return &ACLPlugin{}
}

func (self *ACLPlugin) Name() string {
	return ""
}

func (self *ACLPlugin) Setup(bot *bot.Kabukibot) {
	self.bot = bot
}

func (self *ACLPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &aclPluginWorker{
		bot:     self.bot,
		channel: channel,
	}
}

type aclPluginWorker struct {
	bot     *bot.Kabukibot
	channel bot.Channel
}

func (self *aclPluginWorker) Enable() {
	// nothing to do for us
}

func (self *aclPluginWorker) Disable() {
	// nothing to do for us
}

func (self *aclPluginWorker) Part() {
	// nothing to do for us
}

func (self *aclPluginWorker) Shutdown() {
	// nothing to do for us
}

func (self *aclPluginWorker) Permissions() []string {
	return []string{}
}

var permRegex = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func (self *aclPluginWorker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
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

	self.handleAllowDeny(msg.IsGlobalCommand("allow"), permission, args[1:], sender)
}

var userIdentRegex = regexp.MustCompile(`[^a-zA-Z0-9_$,]`)
var userNameRegex = regexp.MustCompile(`[^a-z0-9_]`)

func (self *aclPluginWorker) handleAllowDeny(allow bool, permission string, args []string, sender bot.Sender) {
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
		sender.Respond("granted permission for " + permission + " to " + strings.Join(processed, ", ") + ".")
	} else {
		sender.Respond("revoked permission for " + permission + " from " + strings.Join(processed, ", ") + ".")
	}
}

func (self *aclPluginWorker) collectPermissions() []string {
	result := make([]string, 0)

	for _, worker := range self.channel.Workers() {
		for _, perm := range worker.Permissions() {
			result = append(result, perm)
		}
	}

	return result
}
