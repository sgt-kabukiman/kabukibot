package monitor

import (
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type worker struct {
	plugin.NilWorker

	config      monitorConfig
	channel     string
	sender      bot.Sender
	sentPing    time.Time
	pending     bool
	delay       time.Duration
	playing     chan struct{}
	stopPlaying chan struct{}
}

func (self *worker) Enable() {
	go self.pingpong()
}

func (self *worker) Disable() {
	close(self.stopPlaying)
	<-self.playing
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if self.pending && strings.ToLower(msg.User.Name) == self.config.ExpectedBy {
		self.pending = false
		self.delay = time.Since(self.sentPing)

		// TODO: Dump the data to a file.
	}
}

func (self *worker) pingpong() {
	defer close(self.playing)

	for {
		select {
		case <-time.After(time.Minute):
			// send ping
			sent := self.sender.SendText(self.config.Message)
			self.pending = true

			// wait for the ping to be sent
			<-sent
			self.sentPing = time.Now()

		case <-self.stopPlaying:
			return
		}
	}
}
