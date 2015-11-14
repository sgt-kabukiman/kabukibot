package plugin

import "strings"
import "regexp"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type ACLPlugin struct {
	bot    *bot.Kabukibot
	acl    *bot.ACL
	prefix string
}

func NewACLPlugin() *ACLPlugin {
	return &ACLPlugin{}
}

func (plugin *ACLPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	plugin.bot = bot
	plugin.acl = bot.ACL()
	plugin.prefix = bot.Configuration().CommandPrefix

	d.OnCommand(plugin.onCommand, nil)
}

var permRegex = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func (plugin *ACLPlugin) onCommand(cmd bot.Command) {
	if cmd.Processed() {
		return
	}

	c := cmd.Command()
	p := plugin.prefix

	// skip unwanted commands
	if c != p+"allow" && c != p+"deny" && c != p+"permissions" && c != p+"allowed" {
		return
	}

	// our commands are all priv only
	user := cmd.User()

	if !user.IsBroadcaster && !plugin.bot.IsOperator(user.Name) {
		return
	}

	channel := cmd.Channel()

	// send the list of available permissions
	if c == p+"permissions" {
		permissions := plugin.getPermissions(channel)

		if len(permissions) == 0 {
			plugin.bot.Say(channel, "There are no permissions available to be configured.")
		} else {
			plugin.bot.Say(channel, "available permissions are: "+strings.Join(permissions, ", "))
		}

		return
	}

	// everything from now on requires at last a permission as the first parameter
	args := cmd.Args()
	if len(args) == 0 {
		plugin.bot.Say(channel, "no permission name given.")
		return
	}

	// check the permission
	permission := strings.ToLower(permRegex.ReplaceAllString(args[0], ""))
	permissions := plugin.getPermissions(channel)

	if len(permission) == 0 {
		plugin.bot.Say(channel, "invalid permission given.")
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
		plugin.bot.Say(channel, "invalid permission ("+permission+") given.")
		return
	}

	// send the list of usernames and groups that have been granted the X permission
	acl := plugin.acl

	if c == p+"allowed" {
		users := acl.AllowedUsers(channel.Name, permission)

		if len(permissions) == 0 {
			plugin.bot.Say(channel, "\""+permission+"\" is granted to nobody at the moment, only you can use it.")
		} else {
			plugin.bot.Say(channel, "\""+permission+"\" is granted to "+strings.Join(users, ", "))
		}

		return
	}

	// no user ident(s) given
	if len(args) == 1 {
		plugin.bot.Say(channel, "no groups/usernames given. Group names are "+strings.Join(bot.ACLGroups(), ", ")+".")
		return
	}

	plugin.handleAllowDeny(c, permission, args[1:], cmd, false)
}

var userIdentRegex = regexp.MustCompile(`[^a-zA-Z0-9_$,]`)
var userNameRegex = regexp.MustCompile(`[^a-z0-9_]`)

func (plugin *ACLPlugin) handleAllowDeny(command string, permission string, args []string, cmd bot.Command, silent bool) {
	// normalize the arguments into a single array of (possibly bogus) idents
	args = strings.Split(strings.ToLower(userIdentRegex.ReplaceAllString(strings.Join(args, ","), "")), ",")

	if len(args) == 0 {
		if !silent {
			plugin.bot.Say(cmd.Channel(), "invalid groups/usernames. Use a comma separated list if you give multiple.")
		}

		return
	}

	isAllow := command == plugin.prefix+"allow"
	processed := make([]string, 0)
	chanName := cmd.Channel().Name

	for i, ident := range args {
		if plugin.acl.IsUsername(ident) {
			ident = userNameRegex.ReplaceAllString(ident, "")

			// if we removed bogus characters, discard the user ident to not accidentally grant or revoke permissions
			if ident != args[i] {
				continue
			}
		}

		if isAllow {
			if plugin.acl.Allow(chanName, ident, permission) {
				processed = append(processed, ident)
			}
		} else {
			if plugin.acl.Deny(chanName, ident, permission) {
				processed = append(processed, ident)
			}
		}
	}

	if silent {
		return
	}

	if len(processed) == 0 {
		plugin.bot.Say(cmd.Channel(), "no changes needed.")
	} else if isAllow {
		plugin.bot.Say(cmd.Channel(), "granted permission for "+permission+" to "+strings.Join(processed, ", ")+".")
	} else {
		plugin.bot.Say(cmd.Channel(), "revoked permission for "+permission+" from "+strings.Join(processed, ", ")+".")
	}
}

func (plugin *ACLPlugin) getPermissions(channel *twitch.Channel) []string {
	plugins := plugin.bot.PluginManager().LoadedPlugins(channel)
	result := make([]string, 0)

	for _, plugin := range plugins {
		for _, perm := range plugin.Permissions() {
			result = append(result, perm)
		}
	}

	return result
}
