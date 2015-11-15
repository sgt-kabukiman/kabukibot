package plugin

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

// TODO: At some point, maybe don't use a mutex for the counter, but some
// channels and a counter goroutine

type SysInfoPlugin struct {
	bot      *bot.Kabukibot
	startup  time.Time
	messages int
	operator string
	mutex    sync.Mutex
}

func NewSysInfoPlugin() *SysInfoPlugin {
	return &SysInfoPlugin{}
}

func (self *SysInfoPlugin) Name() string {
	return ""
}

func (self *SysInfoPlugin) Permissions() []string {
	return []string{}
}

func (self *SysInfoPlugin) Setup(bot *bot.Kabukibot) {
	self.bot = bot
	self.startup = time.Now()
	self.mutex = sync.Mutex{}
	self.operator = bot.OpUsername()
	self.messages = 0
}

func (self *SysInfoPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	return &sysInfoWorker{self}
}

type sysInfoWorker struct {
	plugin *SysInfoPlugin
}

func (self *sysInfoWorker) Enable() {
	// nothing to do for us
}

func (self *sysInfoWorker) Disable() {
	// nothing to do for us
}

func (self *sysInfoWorker) Part() {
	// nothing to do for us
}

func (self *sysInfoWorker) Shutdown() {
	// nothing to do for us
}

func (self *sysInfoWorker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	self.countMessage()

	if msg.IsFrom(self.plugin.operator) {
		if msg.IsGlobalCommand("uptime") {
			sender.SendText("I have been running for " + self.uptime() + ".")
			return
		}

		if msg.IsGlobalCommand("sysinfo") {
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)

			infoString := fmt.Sprintf(
				"System Info: %s uptime, %d channels, %s messages processed, %s res. size",
				self.uptime(), len(self.plugin.bot.Channels()), humanize.FormatInteger("#,###.", self.plugin.messages), humanize.IBytes(mem.Sys),
			)

			sender.SendText(infoString)
		}
	}
}

func (self *sysInfoWorker) HandleClearChatMessage(msg *twitch.ClearChatMessage, sender bot.Sender) {
	self.countMessage()
}

func (self *sysInfoWorker) uptime() string {
	return time.Now().Sub(self.plugin.startup).String()
}

func (self *sysInfoWorker) countMessage() {
	self.plugin.mutex.Lock()
	self.plugin.messages++
	self.plugin.mutex.Unlock()
}
