package plugin

import "strings"
import "regexp"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type BlacklistPlugin struct {
	bot    *bot.Kabukibot
	db     *bot.DatabaseStruct
	log    bot.Logger
	users  []string
	prefix string
}

func NewBlacklistPlugin() *BlacklistPlugin {
	return &BlacklistPlugin{}
}

func (self *BlacklistPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot = bot
	self.db = bot.Database()
	self.log = bot.Logger()
	self.users = make([]string, 0)
	self.prefix = bot.Configuration().CommandPrefix

	self.loadBlacklist()

	d.OnCommand(self.onCommand, nil)
	d.OnTextMessage(self.onTextMessage, nil)
	d.OnTwitchMessage(self.onTwitchMessage, nil)
}

func (self *BlacklistPlugin) onTextMessage(msg twitch.TextMessage) {
	if self.isBlacklisted(msg.User().Name) {
		msg.SetProcessed(true)
	}
}

func (self *BlacklistPlugin) onTwitchMessage(msg twitch.TwitchMessage) {
	if self.isBlacklisted(msg.User().Name) {
		msg.SetProcessed(true)
	}
}

func (self *BlacklistPlugin) onCommand(cmd bot.Command) {
	user := cmd.User()

	if !self.bot.IsOperator(user.Name) {
		return
	}

	// check command

	c := cmd.Command()
	p := self.prefix

	if c != p+"blacklist" && c != p+"unblacklist" {
		return
	}

	// checks args

	args := cmd.Args()

	if len(args) == 0 {
		self.bot.Respond(cmd, "you have to give a username.")
		return
	}

	// sanitise username

	username := args[0]
	cleaner := regexp.MustCompile(`[^a-z0-9]`)

	username = cleaner.ReplaceAllString(strings.ToLower(username), "")

	if len(username) == 0 {
		self.bot.Respond(cmd, "the given username is invalid.")
		return
	}

	// perform blacklisting

	if c == p+"blacklist" {
		if username == user.Name {
			self.bot.Respond(cmd, "you cannot blacklist yourself.")
			return
		}

		if username == self.bot.BotUsername() {
			self.bot.Respond(cmd, "you cannot blacklist me.")
			return
		}

		if self.blacklist(username) {
			self.bot.Respond(cmd, username+" has been blacklisted.")
		} else {
			self.bot.Respond(cmd, username+" is already on the blacklist.")
		}

		return
	}

	// perform unblacklisting

	if self.unblacklist(username) {
		self.bot.Respond(cmd, username+" has been un-blacklisted.")
	} else {
		self.bot.Respond(cmd, username+" is not blacklisted.")
	}
}

func (self *BlacklistPlugin) blacklist(username string) bool {
	if self.isBlacklisted(username) {
		return false
	}

	_, err := self.db.Exec("INSERT INTO blacklist (username) VALUES (?)", username)
	if err != nil {
		self.log.Fatal("Could not insert blacklist entry from the database: " + err.Error())
	}

	self.users = append(self.users, username)

	return true
}

func (self *BlacklistPlugin) unblacklist(username string) bool {
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

func (self *BlacklistPlugin) isBlacklisted(username string) bool {
	for _, u := range self.users {
		if u == username {
			return true
		}
	}

	return false
}

func (self *BlacklistPlugin) loadBlacklist() {
	rows, err := self.db.Query("SELECT username FROM blacklist")
	if err != nil {
		self.log.Fatal("Could not query blacklist: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			self.log.Fatal(err.Error())
		}

		self.users = append(self.users, username)
	}

	if err := rows.Err(); err != nil {
		self.log.Fatal(err.Error())
	}
}
