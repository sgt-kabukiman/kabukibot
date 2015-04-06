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
	acl           *ACL
	chanMngr      *channelManager
	pluginMngr    *pluginManager
	database      *DatabaseStruct
	configuration *Configuration
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

	// hello database
	db := NewDatabase()

	// create the bot
	bot := Kabukibot{}
	bot.configuration = config
	bot.dispatcher    = dispatcher
	bot.twitchClient  = twitchClient
	bot.acl           = NewACL(&bot, db)
	bot.chanMngr      = NewChannelManager(db)
	bot.pluginMngr    = NewPluginManager(&bot, dispatcher, db)
	bot.database      = db

	dispatcher.OnJoin(bot.onJoin)
	dispatcher.OnPart(bot.onPart)
	dispatcher.OnTextMessage(bot.detectCommand)

	return &bot, nil
}

func (bot *Kabukibot) Connect() (chan bool, error) {
	// connect to database
	err := bot.database.Connect(bot.configuration.Database.DSN)
	if err != nil {
		return nil, err
	}

	// setup plugins
	bot.pluginMngr.setup()

	// connect to Twitch
	client := bot.twitchClient

	quitSignal, err := client.Connect()
	if err != nil {
		return nil, err
	}

	// wait for the ready signal, after TWITCHCLIENT has been sent
	<-client.ReadySignal

	// join all of the channels
	bot.joinInitialChannels()

	return quitSignal, nil
}

func (bot *Kabukibot) AddPlugin(plugin Plugin) {
	bot.pluginMngr.registerPlugin(plugin)
}

func (bot *Kabukibot) Dispatcher() Dispatcher {
	return bot.dispatcher
}

func (bot *Kabukibot) ACL() *ACL {
	return bot.acl
}

func (bot *Kabukibot) Configuration() *Configuration {
	return bot.configuration
}

func (bot *Kabukibot) Channels() *channelMap {
	return bot.chanMngr.Channels()
}

func (bot *Kabukibot) Channel(name string) (c *twitch.Channel, ok bool) {
	return bot.chanMngr.Channel(name)
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
	bot.acl.loadChannelData(channel.Name)
	bot.chanMngr.addChannel(channel)
	bot.pluginMngr.setupChannel(channel)
}

func (bot *Kabukibot) onPart(channel *twitch.Channel) {
	bot.pluginMngr.teardownChannel(channel)
	bot.chanMngr.removeChannel(channel)
	bot.acl.unloadChannelData(channel.Name)
}

func (bot *Kabukibot) joinInitialChannels() {
	mngr := bot.chanMngr

	// load all previously joined channels
	mngr.loadChannels()

	// this only needs to be done in an empty database: join ourselves later
	mngr.addChannel(twitch.NewChannel(bot.configuration.Account.Username))

	for _, channel := range *mngr.Channels() {
		bot.Join(channel)
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


