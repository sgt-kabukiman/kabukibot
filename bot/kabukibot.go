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
	logger        Logger
	acl           *ACL
	chanMngr      *channelManager
	pluginMngr    *PluginManager
	dictionary    *Dictionary
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

	// create logger
	logger := NewLogger(LOG_LEVEL_DEBUG)

	// hello database
	db := NewDatabase()

	// create the bot
	bot := Kabukibot{}
	bot.configuration = config
	bot.dispatcher    = dispatcher
	bot.logger        = logger
	bot.twitchClient  = twitchClient
	bot.acl           = NewACL(&bot, logger, db)
	bot.chanMngr      = NewChannelManager(db)
	bot.pluginMngr    = NewPluginManager(&bot, dispatcher, db)
	bot.dictionary    = NewDictionary(db, logger)
	bot.database      = db

	dispatcher.OnJoin(bot.onJoin, nil)
	dispatcher.OnPart(bot.onPart, nil)
	dispatcher.OnTextMessage(bot.detectCommand, nil)

	return &bot, nil
}

func (bot *Kabukibot) Connect() (chan bool, error) {
	// connect to database
	err := bot.database.Connect(bot.configuration.Database.DSN)
	if err != nil {
		return nil, err
	}

	// load dictionary elements
	bot.logger.Debug("Loading dictionary...")
	bot.dictionary.load()

	// setup plugins
	bot.logger.Debug("Setting up plugins...")
	bot.pluginMngr.setup()

	// connect to Twitch
	client := bot.twitchClient

	bot.logger.Info("Connecting to %s:%d...", bot.configuration.IRC.Host, bot.configuration.IRC.Port)
	quitSignal, err := client.Connect()
	if err != nil {
		return nil, err
	}

	// wait for the ready signal, after TWITCHCLIENT has been sent
	<-client.ReadySignal
	bot.logger.Info("Connection established.")

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

func (bot *Kabukibot) Logger() Logger {
	return bot.logger
}

func (bot *Kabukibot) ACL() *ACL {
	return bot.acl
}

func (bot *Kabukibot) Configuration() *Configuration {
	return bot.configuration
}

func (bot *Kabukibot) PluginManager() *PluginManager {
	return bot.pluginMngr
}

func (bot *Kabukibot) Dictionary() *Dictionary {
	return bot.dictionary
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

func (bot *Kabukibot) Respond(msg twitch.Message, text string) {
	bot.twitchClient.Privmsg(msg.Channel().IrcName(), text)
}

func (bot *Kabukibot) RespondToAll(msg twitch.Message, text string) {
	bot.twitchClient.Privmsg(msg.Channel().IrcName(), text)
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


