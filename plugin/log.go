package plugin

// TODO: This is not terribly performant, as it uses unbuffered writes. On high-frequency
// channels, this will slow down a bit, but thankfully only the goroutines for those
// channels.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type logConfig struct {
	Directory string
}

type LogPlugin struct {
	config logConfig
}

func NewLogPlugin() *LogPlugin {
	return &LogPlugin{}
}

func (self *LogPlugin) Name() string {
	return "log"
}

func (self *LogPlugin) Permissions() []string {
	return []string{}
}

func (self *LogPlugin) Setup(bot *bot.Kabukibot) {
	self.config = logConfig{}

	err := bot.Configuration().PluginConfig(self.Name(), &self.config)
	if err != nil {
		bot.Logger().Warn("Could not load 'log' plugin configuration: %s", err)
	}
}

func (self *LogPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &logWorker{
		directory: self.config.Directory,
		channel:   channel.Name(),
	}
}

type logWorker struct {
	directory string
	channel   string
	file      *os.File
}

func (self *logWorker) Enable() {
	self.Disable() // cleanup

	filename := filepath.Join(self.directory, strings.TrimPrefix(self.channel, "#")+".log")

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err == nil {
		self.file = f
	}
}

func (self *logWorker) Disable() {
	if self.file != nil {
		_ = self.file.Close()
		self.file = nil
	}
}

func (self *logWorker) Part() {
	self.Disable()
}

func (self *logWorker) Shutdown() {
	self.Disable()
}

func (self *logWorker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if self.file != nil {
		now := time.Now().Format("2006-Jan-02 15:04:05")
		line := fmt.Sprintf("%s%s: %s", self.userPrefix(msg), msg.User.Name, msg.Text)

		// fmt.Printf("[%s] %s\n", self.channel, line)

		self.file.WriteString(fmt.Sprintf("[%s] %s\n", now, line))
	}
}

func (self *logWorker) HandleClearChatMessage(msg *twitch.ClearChatMessage, sender bot.Sender) {
	if self.file != nil {
		var line string

		now := time.Now().Format("2006-Jan-02 15:04:05")

		if msg.User != "" {
			line = fmt.Sprintf("<%s has been timed out>", msg.User)
		} else {
			line = "<chat has been cleared>"
		}

		// fmt.Printf("[%s] %s\n", self.channel, line)

		self.file.WriteString(fmt.Sprintf("[%s] %s\n", now, line))
	}
}

func (self *logWorker) HandleSubscriberNotificationMessage(msg *twitch.SubscriberNotificationMessage, sender bot.Sender) {
	if self.file != nil {
		now := time.Now().Format("2006-Jan-02 15:04:05")

		// fmt.Printf("[%s] <%s>\n", self.channel, msg.Text)

		self.file.WriteString(fmt.Sprintf("[%s] <%s>\n", now, msg.Text))
	}
}

func (self *logWorker) userPrefix(msg *bot.TextMessage) string {
	prefix := ""
	user := msg.User

	if user.Type == twitch.TwitchAdmin {
		prefix += "!"
	}

	if user.Type == twitch.TwitchStaff {
		prefix += "!!"
	}

	if msg.IsFromOperator() {
		prefix += "$"
	}

	if msg.IsFromBroadcaster() {
		prefix += "&"
	}

	if user.Type == twitch.Moderator {
		prefix += "@"
	}

	if user.Type == twitch.GlobalModerator {
		prefix += "@@"
	}

	if user.Subscriber {
		prefix += "+"
	}

	if user.Turbo {
		prefix += "~"
	}

	return prefix
}
