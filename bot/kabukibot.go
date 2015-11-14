package bot

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type Kabukibot struct {
	twitch       *twitch.TwitchClient
	workers      map[string]*channelWorker
	channelMutex sync.Mutex
	// dispatcher    Dispatcher
	logger Logger
	// acl           *ACL
	// chanMngr      *ChannelManager
	// pluginMngr    *PluginManager
	// dictionary    *Dictionary
	database      *DatabaseStruct
	configuration *Configuration
	alive         chan struct{}
}

func NewKabukibot(config *Configuration) (*Kabukibot, error) {
	// setup our TwitchClient
	server := net.JoinHostPort(config.IRC.Host, strconv.Itoa(config.IRC.Port))
	twitch := twitch.NewTwitchClient(server, config.Account.Username, config.Account.Password, 2*time.Second)

	// we need our own dispatcher to handle custom events
	// dispatcher := NewDispatcher()

	// create logger
	logger := NewLogger(LOG_LEVEL_DEBUG)

	// hello database
	db := NewDatabase()

	// create the bot
	bot := Kabukibot{}
	bot.configuration = config
	bot.workers = make(map[string]*channelWorker)
	bot.channelMutex = sync.Mutex{}
	// bot.dispatcher = dispatcher
	bot.logger = logger
	bot.twitch = twitch
	// bot.acl = NewACL(&bot, logger, db)
	// bot.chanMngr = NewChannelManager(db)
	// bot.pluginMngr = NewPluginManager(&bot, dispatcher, db)
	// bot.dictionary = NewDictionary(db, logger)
	bot.database = db
	bot.alive = make(chan struct{})

	// dispatcher.OnJoin(bot.onJoin, nil)
	// dispatcher.OnPart(bot.onPart, nil)
	// dispatcher.OnTextMessage(bot.detectCommand, nil)

	return &bot, nil
}

func (bot *Kabukibot) Connect() error {
	// connect to database
	err := bot.database.Connect(bot.configuration.Database.DSN)
	if err != nil {
		return err
	}

	// load dictionary elements
	// bot.logger.Debug("Loading dictionary...")
	// bot.dictionary.load()

	// // setup plugins
	// bot.logger.Debug("Setting up plugins...")
	// bot.pluginMngr.setup()

	// connect to Twitch
	client := bot.twitch

	bot.logger.Info("Connecting to %s:%d...", bot.configuration.IRC.Host, bot.configuration.IRC.Port)
	err = client.Connect()
	if err != nil {
		return err
	}

	// wait for the ready signal
	<-client.Ready()
	bot.logger.Info("Connection established.")

	return nil
}

func (bot *Kabukibot) Work() {
	for msg := range bot.twitch.Incoming() {
		// find the appropriate worker
		channel := msg.ChannelName()

		bot.channelMutex.Lock()
		worker, exists := bot.workers[channel]
		bot.channelMutex.Unlock()

		if exists {
			worker.Input() <- msg
		}
	}

	// we're dead now
	close(bot.alive)
}

func (bot *Kabukibot) Alive() <-chan struct{} {
	return bot.alive
}

func (bot *Kabukibot) Configuration() *Configuration {
	return bot.configuration
}

func (bot *Kabukibot) Database() *DatabaseStruct {
	return bot.database
}

func (bot *Kabukibot) Logger() Logger {
	return bot.logger
}

// func (bot *Kabukibot) Dispatcher() Dispatcher {
// 	return bot.dispatcher
// }

// func (bot *Kabukibot) ACL() *ACL {
// 	return bot.acl
// }

// func (bot *Kabukibot) PluginManager() *PluginManager {
// 	return bot.pluginMngr
// }

// func (bot *Kabukibot) ChannelManager() *ChannelManager {
// 	return bot.chanMngr
// }

// func (bot *Kabukibot) EmoteManager() EmoteManager {
// 	return bot.emoteMngr
// }

// func (bot *Kabukibot) Dictionary() *Dictionary {
// 	return bot.dictionary
// }

// func (bot *Kabukibot) Channels() *channelMap {
// 	return bot.chanMngr.Channels()
// }

// func (bot *Kabukibot) Channel(name string) (c *twitch.Channel, ok bool) {
// 	return bot.chanMngr.Channel(name)
// }

// func (bot *Kabukibot) AddPlugin(plugin Plugin) {
// 	bot.pluginMngr.registerPlugin(plugin)
// }

func (bot *Kabukibot) Join(channel string) <-chan bool {
	bot.channelMutex.Lock()

	// check if we already are in the channel. if not, then ..
	_, exists := bot.workers[channel]
	if exists {
		bot.channelMutex.Unlock()

		dummy := make(chan bool, 1)
		dummy <- false
		close(dummy)

		return dummy
	}

	// prepare the channelWorker
	worker := newChannelWorker(channel)

	// remember worker
	bot.workers[channel] = worker
	bot.channelMutex.Unlock()

	// go have fun
	go worker.Work()

	// wait for when the worker dies, which happens when it receives a PART message
	go func() {
		<-worker.Alive()

		// cleanup
		bot.channelMutex.Lock()
		delete(bot.workers, channel)
		bot.channelMutex.Unlock()
	}()

	// now that we are prepared to handle the channel messages, actually join
	return bot.twitch.Send(twitch.JoinMessage{channel})
}

func (bot *Kabukibot) Part(channel string) <-chan bool {
	// send off the request to leave the channel, but wait for its confirmation
	// to shutdown our worker; this signal therefore does not represent the
	// moment when we left, but only the moment the PART message was sent.
	return bot.twitch.Send(twitch.PartMessage{channel})
}

// func (bot *Kabukibot) Respond(msg twitch.Message, text string) {
// 	msg.SetProcessed(true)
// 	bot.twitch.Privmsg(msg.Channel().IrcName(), text)
// }

// func (bot *Kabukibot) RespondToAll(msg twitch.Message, text string) {
// 	msg.SetProcessed(true)
// 	bot.twitch.Privmsg(msg.Channel().IrcName(), text)
// }

// func (bot *Kabukibot) Say(channel *twitch.Channel, text string) {
// 	bot.twitch.Privmsg(channel.IrcName(), text)
// }

func (bot *Kabukibot) BotUsername() string {
	return bot.configuration.Account.Username
}

func (bot *Kabukibot) OpUsername() string {
	return bot.configuration.Operator
}

func (bot *Kabukibot) IsBot(username string) bool {
	return bot.BotUsername() == username
}

func (bot *Kabukibot) IsOperator(username string) bool {
	return bot.OpUsername() == username
}

// func (bot *Kabukibot) onPart(channel *twitch.Channel) {
// 	bot.pluginMngr.teardownChannel(channel)
// 	bot.chanMngr.removeChannel(channel)
// 	bot.acl.unloadChannelData(channel.Name)
// }

func (bot *Kabukibot) joinInitialChannels() {
	// mngr := bot.chanMngr

	// // load all previously joined channels
	// mngr.loadChannels()

	// // this only needs to be done in an empty database: join ourselves later
	// mngr.addChannel(twitch.NewChannel(bot.configuration.Account.Username))

	// for _, channel := range *mngr.Channels() {
	// 	bot.Join(channel)
	// }
}

// var commandRegex = regexp.MustCompile(`^!([a-zA-Z0-9_-]+)(?:\s+(.*))?$`)
// var argSplitter = regexp.MustCompile(`\s+`)

// func (bot *Kabukibot) detectCommand(msg twitch.TextMessage) {
// 	match := commandRegex.FindStringSubmatch(msg.Text())
// 	if len(match) == 0 {
// 		return
// 	}

// 	cmd := strings.ToLower(match[1])
// 	argString := strings.TrimSpace(match[2])
// 	args := make([]string, 0)

// 	if len(argString) > 0 {
// 		args = argSplitter.Split(argString, -1)
// 	}

// 	c := command{msg, cmd, args}
// 	bot.dispatcher.HandleCommand(&c)
// }
