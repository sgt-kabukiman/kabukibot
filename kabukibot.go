package main

import (
	"net"
	"strconv"

	irc "github.com/fluffle/goirc/client"
)

type Kabukibot struct {
	twitchClient  *TwitchClient
	configuration *Configuration
	plugins       []Plugin
	channels      []Channel
}

func NewKabukibot(config *Configuration) (*Kabukibot, error) {
	// setup the IRC client

	cfg := irc.NewConfig(config.Account.Username)

	cfg.SSL     = false
	cfg.Pass    = config.Account.Password
	cfg.Server  = net.JoinHostPort(config.IRC.Host, strconv.Itoa(config.IRC.Port))
	cfg.NewNick = func(n string) string { return n + "^" }
	cfg.Flood   = true // means no flood protection

	ircClient := irc.Client(cfg)

	// setup our TwitchClient wrapper

	twitchClient := NewTwitchClient(ircClient)

	// create the bot

	bot := Kabukibot{}
	bot.configuration = config
	bot.twitchClient  = twitchClient
	bot.plugins       = make([]Plugin, 10)
	bot.channels      = make([]Channel, 10)

	return &bot, nil
}

func (bot *Kabukibot) Connect() (chan bool, error) {
	return bot.twitchClient.Connect()
}

func (bot *Kabukibot) AddPlugin(plugin Plugin) {
	bot.plugins = append(bot.plugins, plugin)
}

func (bot *Kabukibot) On(event string, fn irc.HandlerFunc) irc.Remover {
	return bot.twitchClient.On(event, fn)
}

func (bot *Kabukibot) OnBG(event string, fn irc.HandlerFunc) irc.Remover {
	return bot.twitchClient.OnBG(event, fn)
}
