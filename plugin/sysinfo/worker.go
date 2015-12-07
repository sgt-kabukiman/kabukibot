package sysinfo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

func (self *pluginStruct) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	self.countMessage()

	if msg.IsFromOperator() {
		if msg.IsGlobalCommand("uptime") {
			sender.Respond("I have been running for " + self.uptime() + ".")
			return
		}

		if msg.IsGlobalCommand("sysinfo") {
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)

			infoString := fmt.Sprintf(
				"System Info: %s uptime, %d channels, %s messages processed, %s res. size",
				self.uptime(), len(self.bot.Channels()), humanize.FormatInteger("#,###.", self.messages), humanize.IBytes(mem.Sys),
			)

			sender.Respond(infoString)
		}
	}
}

func (self *pluginStruct) HandleClearChatMessage(msg *twitch.ClearChatMessage, sender bot.Sender) {
	self.countMessage()
}

func (self *pluginStruct) uptime() string {
	return time.Now().Sub(self.startup).String()
}

func (self *pluginStruct) countMessage() {
	self.mutex.Lock()
	self.messages++
	self.mutex.Unlock()
}
