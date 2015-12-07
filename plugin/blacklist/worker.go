package blacklist

import (
	"regexp"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

func (self *pluginStruct) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if msg.IsProcessed() {
		return
	}

	// mark messages from blacklisted users as processed
	if self.isBlacklisted(msg.User.Name) {
		self.log.Info("User is blacklisted.")
		msg.SetProcessed()
		return
	}

	if !msg.IsFromOperator() {
		return
	}

	if !msg.IsGlobalCommand("blacklist") && !msg.IsGlobalCommand("unblacklist") {
		return
	}

	// checks args

	args := msg.Arguments()

	if len(args) == 0 {
		sender.Respond("you have to give a username.")
		return
	}

	// sanitise username

	cleaner := regexp.MustCompile(`[^a-z0-9_]`)
	username := cleaner.ReplaceAllString(strings.ToLower(args[0]), "")

	if len(username) == 0 {
		sender.Respond("the given username is invalid.")
		return
	}

	// perform blacklisting

	if msg.IsGlobalCommand("blacklist") {
		if username == strings.ToLower(msg.User.Name) {
			sender.Respond("you cannot blacklist yourself.")
			return
		}

		if username == self.bot {
			sender.Respond("you cannot blacklist me.")
			return
		}

		if self.blacklist(username) {
			sender.Respond(username + " has been blacklisted.")
		} else {
			sender.Respond(username + " is already on the blacklist.")
		}

		return
	}

	// perform unblacklisting

	if self.unblacklist(username) {
		sender.Respond(username + " has been un-blacklisted.")
	} else {
		sender.Respond(username + " is not blacklisted.")
	}
}

func (self *pluginStruct) blacklist(username string) bool {
	// use a read-lock around the isBlacklisted check
	self.mutex.RLock()

	if self.isBlacklisted(username) {
		self.mutex.RUnlock()
		return false
	}

	self.mutex.RUnlock()
	self.mutex.Lock()

	_, err := self.db.Exec("INSERT INTO blacklist (username) VALUES (?)", username)
	if err != nil {
		self.log.Fatal("Could not insert blacklist entry from the database: " + err.Error())
	}

	self.users = append(self.users, username)
	self.mutex.Unlock()

	return true
}

func (self *pluginStruct) unblacklist(username string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	pos := -1

	for idx, u := range self.users {
		if u == username {
			pos = idx
			break
		}
	}

	if pos == -1 {
		return false
	}

	_, err := self.db.Exec("DELETE FROM blacklist WHERE username = ?", username)
	if err != nil {
		self.log.Fatal("Could not delete blacklist entry from the database: " + err.Error())
	}

	self.users = append(self.users[:pos], self.users[(pos+1):]...)

	return true
}

func (self *pluginStruct) isBlacklisted(username string) bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	for _, u := range self.users {
		if u == username {
			return true
		}
	}

	return false
}

type blacklistUser struct {
	Username string
}

func (self *pluginStruct) loadBlacklist() {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	list := make([]blacklistUser, 0)
	self.db.Select(&list, "SELECT username FROM blacklist ORDER BY username")

	self.users = make([]string, 0)

	for _, u := range list {
		self.users = append(self.users, u.Username)
	}
}
