package bot

import "log"
import "strings"
import "github.com/sgt-kabukiman/kabukibot/twitch"

const (
	ACL_ALL           = "$all"
	ACL_MODERATORS    = "$mods"
	ACL_SUBSCRIBERS   = "$subs"
	ACL_TURBO_USERS   = "$turbos"
	ACL_TWITCH_STAFF  = "$staff"
	ACL_TWITCH_ADMINS = "$admins"
)

type usernameList []string
type permissionMap map[string]usernameList
type channelPermMap map[string]permissionMap

type ACL struct {
	bot         *Kabukibot
	log         Logger
	db          *DatabaseStruct
	permissions channelPermMap
}

func NewACL(bot *Kabukibot, log Logger, db *DatabaseStruct) *ACL {
	return &ACL{bot, log, db, make(channelPermMap)}
}

func ACLGroups() []string {
	return []string{ACL_ALL, ACL_MODERATORS, ACL_SUBSCRIBERS, ACL_TURBO_USERS, ACL_TWITCH_STAFF, ACL_TWITCH_ADMINS}
}

func (self *ACL) AllowedUsers(channel string, permission string) *usernameList {
	permMap, ok := self.permissions[channel]
	if !ok {
		empty := make(usernameList, 0)
		return &empty
	}

	userList, ok := permMap[permission]
	if !ok {
		empty := make(usernameList, 0)
		return &empty
	}

	return &userList
}

func (self *ACL) IsUsername(name string) bool {
	name = strings.ToLower(name)

	for _, group := range ACLGroups() {
		if group == name {
			return false
		}
	}

	return true
}

func (self *ACL) IsAllowed(user *twitch.User, permission string) bool {
	// the bot operator and channel owner are always allowed to use the available commands
	if user.IsBroadcaster || self.bot.IsOperator(user.Name) {
		return true
	}

	allowed := self.AllowedUsers(user.Channel.Name, permission)
	if len(*allowed) == 0 {
		return false
	}

	for _, ident := range *allowed {
		allowed := false

		switch ident {
		case ACL_ALL:
			allowed = true
		case ACL_MODERATORS:
			allowed = user.IsModerator
		case ACL_SUBSCRIBERS:
			allowed = user.IsSubscriber
		case ACL_TURBO_USERS:
			allowed = user.IsTurbo
		case ACL_TWITCH_STAFF:
			allowed = user.IsTwitchStaff
		case ACL_TWITCH_ADMINS:
			allowed = user.IsTwitchAdmin
		default:
			allowed = user.Name == ident
		}

		if allowed {
			return true
		}
	}

	return false
}

func (self *ACL) Allow(channel string, userIdent string, permission string) bool {
	// allowing something for the owner is pointless
	if channel == userIdent {
		return false
	}

	// create skeleton structure for channel and permissions
	_, ok := self.permissions[channel]
	if !ok {
		self.permissions[channel] = make(permissionMap, 0)
	}

	_, ok = self.permissions[channel][permission]
	if !ok {
		self.permissions[channel][permission] = make(usernameList, 0)
	}

	// check if the permission is already set
	exists := false

	for _, ident := range self.permissions[channel][permission] {
		if ident == userIdent {
			exists = true
			break
		}
	}

	if exists {
		return false
	}

	self.permissions[channel][permission] = append(self.permissions[channel][permission], userIdent)

	_, err := self.db.Exec("INSERT INTO acl (channel, permission, user_ident) VALUES (?,?,?)", channel, permission, userIdent)
	if err != nil {
		log.Fatal("Could not add ACL entry to the database: " + err.Error())
	}

	self.log.Debug("Allowed %s for %s in #%s.", permission, userIdent, channel)

	return true
}

func (self *ACL) Deny(channel string, userIdent string, permission string) bool {
	permMap, ok := self.permissions[channel]
	if !ok {
		return false
	}

	userList, ok := permMap[permission]
	if !ok {
		return false
	}

	idx := -1

	for i, ident := range userList {
		if ident == userIdent {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false
	}

	// remove element or kill list alltogether if this was the last user
	if len(userList) > 1 {
		self.permissions[channel][permission] = append(userList[:idx], userList[(idx+1):]...)
	} else {
		delete(self.permissions[channel], permission)

		// kill the entire channel if it's empty now
		if len(self.permissions[channel]) == 0 {
			delete(self.permissions, channel)
		}
	}

	_, err := self.db.Exec("DELETE FROM acl WHERE channel = ? AND permission = ? AND user_ident = ?", channel, permission, userIdent)
	if err != nil {
		log.Fatal("Could not delete ACL entry from the database: " + err.Error())
	}

	self.log.Debug("Denied %s for %s in #%s.", permission, userIdent, channel)

	return true
}

func (self *ACL) DeletePermission(channel string, permission string) {
	permMap, ok := self.permissions[channel]
	if !ok {
		return
	}

	_, ok = permMap[permission]
	if !ok {
		return
	}

	delete(self.permissions[channel], permission)

	// kill the entire channel if it's empty now
	if len(self.permissions[channel]) == 0 {
		delete(self.permissions, channel)
	}

	_, err := self.db.Exec("DELETE FROM acl WHERE channel = ? AND permission = ?", channel, permission)
	if err != nil {
		log.Fatal("Could not delete ACL entries from the database: " + err.Error())
	}

	self.log.Debug("Removed all %s permissions for #%s.", permission, channel)
}

func (self *ACL) loadChannelData(channel string) {
	rows, err := self.db.Query("SELECT permission, user_ident FROM acl WHERE channel = ?", channel)
	if err != nil {
		self.log.Fatal("Could not query ACL data: %s", err.Error())
	}
	defer rows.Close()

	newPermMap := make(permissionMap)
	lastPerm := ""
	rowCount := 0

	var newUserList usernameList

	for rows.Next() {
		var permission, userIdent string

		if err := rows.Scan(&permission, &userIdent); err != nil {
			log.Fatal(err)
		}

		if permission != lastPerm {
			if lastPerm != "" {
				newPermMap[lastPerm] = newUserList
			}

			newUserList = make(usernameList, 0)
			lastPerm = permission
		}

		newUserList = append(newUserList, userIdent)
		rowCount = rowCount + 1
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if lastPerm != "" {
		newPermMap[lastPerm] = newUserList
	}

	self.permissions[channel] = newPermMap

	self.log.Debug("Loaded %d ACL entries for #%s.", rowCount, channel)
}

func (self *ACL) unloadChannelData(channel string) {
	_, ok := self.permissions[channel]
	if ok {
		delete(self.permissions, channel)
		self.log.Debug("Unloaded ACL data for #%s.", channel)
	}
}
