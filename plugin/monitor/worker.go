package monitor

import (
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type worker struct {
	plugin.NilWorker

	log         bot.Logger
	bot         *bot.Kabukibot
	startup     time.Time
	config      monitorConfig
	channel     string
	sender      bot.Sender
	sentPing    time.Time
	pending     bool
	delay       time.Duration
	playing     chan struct{}
	stopPlaying chan struct{}
	dumping     chan struct{}
	stopDumping chan struct{}
	sent        uint64
	received    uint64
}

func (self *worker) Enable() {
	go self.pingpong()
	go self.dumper()
}

func (self *worker) Disable() {
	close(self.stopPlaying)
	<-self.playing

	close(self.stopDumping)
	<-self.dumping
}

type monitorStatus struct {
	Uptime   string `json:"uptime"`
	Channels int    `json:"channels"`
	Memory   struct {
		Residential uint64 `json:"rss"`
		HeapTotal   uint64 `json:"heapTotal"`
		HeapUsed    uint64 `json:"heapUsed"`
	} `json:"memory"`
	Messages struct {
		Received uint64 `json:"received"`
		Sent     uint64 `json:"sent"`
	} `json:"messages"`
	Queue     int `json:"queue"`
	Heartbeat int `json:"heartbeat"`
}

func (self *worker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if self.pending && strings.ToLower(msg.User.Name) == self.config.ExpectedBy {
		self.pending = false
		self.delay = time.Since(self.sentPing)
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

func (self *worker) dumper() {
	defer close(self.dumping)

	for {
		select {
		case <-time.After(time.Minute):
			memStats := runtime.MemStats{}
			runtime.ReadMemStats(&memStats)

			s := self.bot.MessagesSent()
			r := self.bot.MessagesReceived()

			status := monitorStatus{}
			status.Uptime = time.Since(self.startup).String()
			status.Channels = len(self.bot.Channels())
			status.Memory.Residential = memStats.Sys
			status.Memory.HeapTotal = memStats.HeapSys
			status.Memory.HeapUsed = memStats.HeapInuse
			status.Messages.Received = r - self.received
			status.Messages.Sent = s - self.sent
			status.Queue = self.bot.QueueLen()
			status.Heartbeat = int(self.delay.Nanoseconds() / int64(time.Millisecond))

			file, err := os.OpenFile(self.config.Filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
			if err != nil {
				self.log.Error("Could not open monitor status file: %s", err.Error())
			} else {
				encoder := json.NewEncoder(file)

				if err := encoder.Encode(&status); err != nil {
					self.log.Error("Could not dump monitor status: %s", err.Error())
				}

				file.Close()
			}

			self.sent = s
			self.received = r

		case <-self.stopDumping:
			return
		}
	}
}
