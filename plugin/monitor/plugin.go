package monitor

import (
	"strings"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

type monitorConfig struct {
	Channel    string
	Message    string
	ExpectedBy string `yaml:"expectedBy"`
	Filename   string
}

type pluginStruct struct {
	plugin.BasePlugin

	config  monitorConfig
	startup time.Time
	bot     *bot.Kabukibot
}

func NewPlugin() *pluginStruct {
	return &pluginStruct{
		startup: time.Now(),
	}
}

func (self *pluginStruct) Setup(bot *bot.Kabukibot) {
	self.config = monitorConfig{}
	self.bot = bot

	err := bot.Configuration().PluginConfig("monitor", &self.config)
	if err != nil {
		bot.Logger().Warning("Could not load 'monitor' plugin configuration: %s", err)
	}

	self.config.ExpectedBy = strings.ToLower(self.config.ExpectedBy)
}

func (self *pluginStruct) CreateWorker(channel bot.Channel) bot.PluginWorker {
	if channel.Name() == self.config.Channel {
		return &worker{
			bot:         self.bot,
			log:         self.bot.Logger(),
			startup:     self.startup,
			config:      self.config,
			channel:     channel.Name(),
			sender:      channel.Sender(),
			playing:     make(chan struct{}),
			stopPlaying: make(chan struct{}),
			dumping:     make(chan struct{}),
			stopDumping: make(chan struct{}),
		}
	} else {
		return &plugin.NilWorker{}
	}
}
