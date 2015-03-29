package bot

import (
	"net"
	"fmt"
	"strconv"

	irc "github.com/fluffle/goirc/client"
	logging "github.com/fluffle/goirc/logging"
)

type Kabukibot struct {
	twitchClient  *TwitchClient
	dispatcher    *Dispatcher
	configuration *Configuration
	plugins       []Plugin
	channels      []Channel
}

type debugLogger struct{}

func (log *debugLogger) Debug(format string, args ...interface{}) { fmt.Printf("[DBG] " + format + "\n", args...) }
func (log *debugLogger) Info(format string, args ...interface{})  { fmt.Printf("[INF] " + format + "\n", args...) }
func (log *debugLogger) Warn(format string, args ...interface{})  { fmt.Printf("[WRN] " + format + "\n", args...) }
func (log *debugLogger) Error(format string, args ...interface{}) { fmt.Printf("[ERR] " + format + "\n", args...) }

func NewKabukibot(config *Configuration) (*Kabukibot, error) {
	// log everything
	logging.SetLogger(&debugLogger{})

	// setup the IRC client
	cfg := irc.NewConfig(config.Account.Username)

	cfg.SSL     = false
	cfg.Pass    = config.Account.Password
	cfg.Server  = net.JoinHostPort(config.IRC.Host, strconv.Itoa(config.IRC.Port))
	cfg.NewNick = func(n string) string { return n + "^" }
	cfg.Flood   = true // means no flood protection

	ircClient := irc.Client(cfg)

	// we need an event dispatcher
	dispatcher := NewDispatcher()

	// setup our TwitchClient wrapper
	twitchClient := NewTwitchClient(ircClient, dispatcher)

	// create the bot
	bot := Kabukibot{}
	bot.configuration = config
	bot.dispatcher    = dispatcher
	bot.twitchClient  = twitchClient
	bot.plugins       = make([]Plugin, 0)
	bot.channels      = make([]Channel, 0)

	return &bot, nil
}

func (bot *Kabukibot) Connect() (chan bool, error) {
	// setup plugins
	for _, plugin := range bot.plugins {
		plugin.Setup(bot, bot.Dispatcher())
	}

	client := bot.twitchClient

	quitChan, err := client.Connect()
	if err != nil {
		return nil, err
	}

	// wait for the ready signal, after TWITCHCLIENT has been sent
	<-client.ready
	client.conn.Join("#kabukibot")
	client.conn.Join("#kabukibotdev")

	return quitChan, nil
}

func (bot *Kabukibot) AddPlugin(plugin Plugin) {
	bot.plugins = append(bot.plugins, plugin)
}

func (bot *Kabukibot) Dispatcher() *Dispatcher {
	return bot.dispatcher
}

func (bot *Kabukibot) Configuration() *Configuration {
	return bot.configuration
}
