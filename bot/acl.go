package bot

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)
import "strings"

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

type ACL struct {
	channel     string
	operator    string
	broadcaster string
	log         Logger
	db          *sqlx.DB
	permissions permissionMap
}

func NewACL(channel string, operator string, log Logger, db *sqlx.DB) *ACL {
	return &ACL{channel, strings.ToLower(operator), strings.ToLower(strings.TrimPrefix(operator, "#")), log, db, make(permissionMap)}
}

func ACLGroups() []string {
	return []string{ACL_ALL, ACL_MODERATORS, ACL_SUBSCRIBERS, ACL_TURBO_USERS, ACL_TWITCH_STAFF, ACL_TWITCH_ADMINS}
}

func (self *ACL) AllowedUsers(permission string) usernameList {
	userList, ok := self.permissions[permission]
	if !ok {
		userList = make(usernameList, 0)
	}

	return userList
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

func (self *ACL) IsAllowed(user twitch.User, permission string) bool {
	name := strings.ToLower(user.Name)

	// the bot operator and channel owner are always allowed to use the available commands
	if name == self.operator || name == self.broadcaster {
		return true
	}

	allowed := self.AllowedUsers(permission)
	if len(allowed) == 0 {
		return false
	}

	for _, ident := range allowed {
		allowed := false

		switch ident {
		case ACL_ALL:
			allowed = true
		case ACL_MODERATORS:
			allowed = (user.Type == twitch.Moderator) || (user.Type == twitch.GlobalModerator)
		case ACL_SUBSCRIBERS:
			allowed = user.Subscriber
		case ACL_TURBO_USERS:
			allowed = user.Turbo
		case ACL_TWITCH_STAFF:
			allowed = user.Type == twitch.TwitchStaff
		case ACL_TWITCH_ADMINS:
			allowed = user.Type == twitch.TwitchAdmin
		default:
			allowed = name == ident
		}

		if allowed {
			return true
		}
	}

	return false
}

func (self *ACL) Allow(userIdent string, permission string) bool {
	userIdent = strings.ToLower(userIdent)

	// allowing something for the owner is pointless
	if self.broadcaster == userIdent {
		return false
	}

	// create skeleton structure for permissions
	_, ok := self.permissions[permission]
	if !ok {
		self.permissions[permission] = make(usernameList, 0)
	}

	// check if the permission is already set
	exists := false

	for _, ident := range self.permissions[permission] {
		if ident == userIdent {
			exists = true
			break
		}
	}

	if exists {
		return false
	}

	self.permissions[permission] = append(self.permissions[permission], userIdent)

	_, err := self.db.Exec("INSERT INTO acl (channel, permission, user_ident) VALUES (?,?,?)", self.channel, permission, userIdent)
	if err != nil {
		log.Fatal("Could not add ACL entry to the database: " + err.Error())
	}

	self.log.Debug("Allowed %s for %s in %s.", permission, userIdent, self.channel)

	return true
}

func (self *ACL) Deny(userIdent string, permission string) bool {
	userIdent = strings.ToLower(userIdent)

	userList, ok := self.permissions[permission]
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
		self.permissions[permission] = append(userList[:idx], userList[(idx+1):]...)
	} else {
		delete(self.permissions, permission)
	}

	_, err := self.db.Exec("DELETE FROM acl WHERE channel = ? AND permission = ? AND user_ident = ?", self.channel, permission, userIdent)
	if err != nil {
		log.Fatal("Could not delete ACL entry from the database: " + err.Error())
	}

	self.log.Debug("Denied %s for %s in %s.", permission, userIdent, self.channel)

	return true
}

func (self *ACL) DeletePermission(permission string) {
	_, ok := self.permissions[permission]
	if !ok {
		return
	}

	delete(self.permissions, permission)

	_, err := self.db.Exec("DELETE FROM acl WHERE channel = ? AND permission = ?", self.channel, permission)
	if err != nil {
		log.Fatal("Could not delete ACL entries from the database: " + err.Error())
	}

	self.log.Debug("Removed all %s permissions for %s.", permission, self.channel)
}

func (self *ACL) loadData() {
	rows, err := self.db.Query("SELECT permission, user_ident FROM acl WHERE channel = ? ORDER BY permission", self.channel)
	if err != nil {
		self.log.Fatal("Could not query ACL data: %s", err.Error())
	}
	defer rows.Close()

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
				self.permissions[lastPerm] = newUserList
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
		self.permissions[lastPerm] = newUserList
	}

	self.log.Debug("Loaded %d ACL entries for %s.", rowCount, self.channel)
}
