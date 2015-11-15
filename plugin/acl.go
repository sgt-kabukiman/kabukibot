package plugin

import (
	"regexp"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type ACLPlugin struct {
	bot      *bot.Kabukibot
	operator string
}

func NewACLPlugin() *ACLPlugin {
	return &ACLPlugin{}
}

func (self *ACLPlugin) Name() string {
	return ""
}

func (self *ACLPlugin) Permissions() []string {
	return []string{}
}

func (self *ACLPlugin) Setup(bot *bot.Kabukibot) {
	self.bot = bot
	self.operator = bot.OpUsername()
}

func (self *ACLPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &aclPluginWorker{
		bot:      self.bot,
		operator: self.operator,
		channel:  channel,
	}
}

type aclPluginWorker struct {
	bot      *bot.Kabukibot
	operator string
	channel  bot.Channel
}

func (self *aclPluginWorker) Part() {
	// nothing to do for us
}

func (self *aclPluginWorker) Shutdown() {
	// nothing to do for us
}

var permRegex = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func (self *aclPluginWorker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	// skip unwanted commands
	if !msg.IsGlobalCommand("allow") && !msg.IsGlobalCommand("deny") && !msg.IsGlobalCommand("permissions") && !msg.IsGlobalCommand("allowed") {
		return
	}

	// our commands are all priv-only
	if !msg.IsFromBroadcaster() && !msg.IsFrom(self.operator) {
		return
	}

	// send the list of available permissions
	if msg.IsGlobalCommand("permissions") {
		permissions := self.permissions()

		if len(permissions) == 0 {
			sender.SendText("There are no permissions available to be configured.")
		} else {
			sender.SendText("available permissions are: " + strings.Join(permissions, ", "))
		}

		return
	}

	// everything from now on requires at last a permission as the first parameter
	args := msg.Arguments()
	if len(args) == 0 {
		sender.SendText("no permission name given.")
		return
	}

	// check the permission
	permission := strings.ToLower(permRegex.ReplaceAllString(args[0], ""))
	permissions := self.permissions()

	if len(permission) == 0 {
		sender.SendText("invalid permission given.")
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
		sender.SendText("invalid permission (" + permission + ") given.")
		return
	}

	// send the list of usernames and groups that have been granted the X permission
	acl := self.channel.ACL()

	if msg.IsGlobalCommand("allowed") {
		users := acl.AllowedUsers(permission)

		if len(permissions) == 0 {
			sender.SendText("\"" + permission + "\" is granted to nobody at the moment, only you can use it.")
		} else {
			sender.SendText("\"" + permission + "\" is granted to " + strings.Join(users, ", "))
		}

		return
	}

	// no user ident(s) given
	if len(args) == 1 {
		sender.SendText("no groups/usernames given. Group names are " + strings.Join(bot.ACLGroups(), ", ") + ".")
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
		sender.SendText("invalid groups/usernames. Use a comma separated list if you give multiple.")
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
		sender.SendText("no changes needed.")
	} else if allow {
		sender.SendText("granted permission for " + permission + " to " + strings.Join(processed, ", ") + ".")
	} else {
		sender.SendText("revoked permission for " + permission + " from " + strings.Join(processed, ", ") + ".")
	}
}

func (self *aclPluginWorker) permissions() []string {
	result := make([]string, 0)

	for _, plugin := range self.channel.Plugins() {
		for _, perm := range plugin.Permissions() {
			result = append(result, perm)
		}
	}

	return result
}
