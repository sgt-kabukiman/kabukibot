package bot

import (
	"time"
	"net"
	"strconv"
	"strings"
	"regexp"

	irc "github.com/fluffle/goirc/client"
	// logging "github.com/fluffle/goirc/logging"
	twitch "github.com/sgt-kabukiman/kabukibot/twitch"
)

type Kabukibot struct {
	twitchClient  *twitch.TwitchClient
	dispatcher    Dispatcher
	configuration *Configuration
	plugins       []Plugin
}

func NewKabukibot(config *Configuration) (*Kabukibot, error) {
	// log everything
	// logging.SetLogger(&debugLogger{})

	// setup the IRC client
	cfg := irc.NewConfig(config.Account.Username)

	cfg.SSL     = false
	cfg.Pass    = config.Account.Password
	cfg.Server  = net.JoinHostPort(config.IRC.Host, strconv.Itoa(config.IRC.Port))
	cfg.NewNick = func(n string) string { return n + "^" }
	cfg.Flood   = true // means no flood protection

	ircClient := irc.Client(cfg)

	// we need our own dispatcher to handle custom events
	dispatcher := NewDispatcher()

	// setup our TwitchClient wrapper
	twitchClient := twitch.NewTwitchClient(ircClient, dispatcher, 2*time.Second)

	// create the bot
	bot := Kabukibot{}
	bot.configuration = config
	bot.dispatcher    = dispatcher
	bot.twitchClient  = twitchClient
	bot.plugins       = make([]Plugin, 0)

	dispatcher.OnJoin(bot.onJoin)
	dispatcher.OnPart(bot.onPart)
	dispatcher.OnTextMessage(bot.detectCommand)

	return &bot, nil
}

func (bot *Kabukibot) Connect() (chan bool, error) {
	// setup plugins
	for _, plugin := range bot.plugins {
		plugin.Setup(bot, bot.Dispatcher())
	}

	client := bot.twitchClient

	quitSignal, err := client.Connect()
	if err != nil {
		return nil, err
	}

	// wait for the ready signal, after TWITCHCLIENT has been sent
	<-client.ReadySignal

	return quitSignal, nil
}

func (bot *Kabukibot) AddPlugin(plugin Plugin) {
	bot.plugins = append(bot.plugins, plugin)
}

func (bot *Kabukibot) Dispatcher() Dispatcher {
	return bot.dispatcher
}

func (bot *Kabukibot) Configuration() *Configuration {
	return bot.configuration
}

func (bot *Kabukibot) Channel(name string) (c *twitch.Channel, ok bool) {
	return bot.twitchClient.Channel(name)
}

func (bot *Kabukibot) Join(channel *twitch.Channel) {
	bot.twitchClient.Join(channel)
}

func (bot *Kabukibot) Part(channel *twitch.Channel) {
	bot.twitchClient.Part(channel)
}

func (bot *Kabukibot) Say(channel *twitch.Channel, text string) {
	bot.twitchClient.Privmsg(channel.IrcName(), text)
}

func (bot *Kabukibot) IsBot(username string) bool {
	return bot.configuration.Account.Username == username
}

func (bot *Kabukibot) IsOperator(username string) bool {
	return bot.configuration.Operator == username
}

func (bot *Kabukibot) onJoin(channel *twitch.Channel) {
	for _, plugin := range bot.plugins {
		switch p := plugin.(type) {
		case ChannelPlugin:
			p.Load(channel, bot, bot.dispatcher)
		}
	}
}

func (bot *Kabukibot) onPart(channel *twitch.Channel) {
	for _, plugin := range bot.plugins {
		switch p := plugin.(type) {
		case ChannelPlugin:
			p.Unload(channel, bot, bot.dispatcher)
		}
	}
}

var commandRegex = regexp.MustCompile(`^!([a-zA-Z0-9_-]+)(?:\s+(.*))?$`)
var argSplitter  = regexp.MustCompile(`\s+`)

func (bot *Kabukibot) detectCommand(msg twitch.TextMessage) {
	match := commandRegex.FindStringSubmatch(msg.Text())
	if len(match) == 0 {
		return
	}

	cmd       := strings.ToLower(match[1])
	argString := strings.TrimSpace(match[2])
	args      := make([]string, 0)

	if len(argString) > 0 {
		args = argSplitter.Split(argString, -1)
	}

	c := command{msg, cmd, args}
	bot.dispatcher.HandleCommand(&c)
}


