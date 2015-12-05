package plugin

import (
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
)

type monitorConfig struct {
	Channel    string
	Message    string
	ExpectedBy string `yaml:"expectedBy"`
	Filename   string
}

type MonitorPlugin struct {
	config monitorConfig
}

func NewMonitorPlugin() *MonitorPlugin {
	return &MonitorPlugin{}
}

func (self *MonitorPlugin) Name() string {
	return ""
}

func (self *MonitorPlugin) Setup(bot *bot.Kabukibot) {
	self.config = monitorConfig{}

	err := bot.Configuration().PluginConfig("monitor", &self.config)
	if err != nil {
		bot.Logger().Warn("Could not load 'monitor' plugin configuration: %s", err)
	}

	self.config.ExpectedBy = strings.ToLower(self.config.ExpectedBy)
}

func (self *MonitorPlugin) CreateWorker(channel bot.Channel) bot.PluginWorker {
	if channel.Name() == self.config.Channel {
		return &monitorWorker{
			config:      self.config,
			channel:     channel.Name(),
			sender:      channel.Sender(),
			playing:     make(chan struct{}),
			stopPlaying: make(chan struct{}),
		}
	} else {
		return &nilWorker{}
	}
}

type monitorWorker struct {
	nilWorker

	config      monitorConfig
	channel     string
	sender      bot.Sender
	sentPing    time.Time
	pending     bool
	delay       time.Duration
	playing     chan struct{}
	stopPlaying chan struct{}
}

func (self *monitorWorker) Enable() {
	go self.pingpong()
}

func (self *monitorWorker) Disable() {
	close(self.stopPlaying)
	<-self.playing
}

func (self *monitorWorker) HandleTextMessage(msg *bot.TextMessage, sender bot.Sender) {
	if self.pending && strings.ToLower(msg.User.Name) == self.config.ExpectedBy {
		self.pending = false
		self.delay = time.Since(self.sentPing)

		// TODO: Dump the data to a file.
	}
}

func (self *monitorWorker) pingpong() {
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
