package plugin

import "fmt"
import "time"
import "runtime"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type SysInfoPlugin struct {
	bot      *bot.Kabukibot
	chans    *bot.ChannelManager
	plugins  *bot.PluginManager
	startup  time.Time
	messages uint64
	prefix   string
}

func NewSysInfoPlugin() *SysInfoPlugin {
	return &SysInfoPlugin{}
}

func (self *SysInfoPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	self.bot      = bot
	self.chans    = bot.ChannelManager()
	self.plugins  = bot.PluginManager()
	self.prefix   = bot.Configuration().CommandPrefix
	self.startup  = time.Now()
	self.messages = 0

	d.OnCommand(self.onCommand, nil)
	d.OnTextMessage(self.onTextMessage, nil)
	d.OnTwitchMessage(self.onTwitchMessage, nil)
}

func (self *SysInfoPlugin) onTextMessage(m twitch.TextMessage) {
	self.messages++
}

func (self *SysInfoPlugin) onTwitchMessage(m twitch.TwitchMessage) {
	if m.Command() == "clearchat" {
		self.messages++
	}
}

func (self *SysInfoPlugin) onCommand(cmd bot.Command) {
	if cmd.Processed() || !self.bot.IsOperator(cmd.User().Name) {
		return
	}

	command := cmd.Command()
	prefix  := self.prefix

	if command == prefix+"uptime" {
		self.bot.RespondToAll(cmd, "I have been running for " + self.getUptime() + ".")
		return
	}

	if command == prefix+"sysinfo" {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		infoString := fmt.Sprintf(
			"System Info: %s uptime, %d channels, %d messages processed, %.2f MiB res. size",
			self.getUptime(), len(*self.bot.Channels()), self.messages, float64(mem.Sys) / (1024*1024),
		)

		self.bot.RespondToAll(cmd, infoString)
	}
}

func (self *SysInfoPlugin) getUptime() string {
	return time.Now().Sub(self.startup).String()
}
