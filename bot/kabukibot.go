package bot

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type Kabukibot struct {
	twitch       *twitch.TwitchClient
	workers      map[string]*channelWorker
	channelMutex sync.Mutex
	plugins      []Plugin
	logger       Logger
	// chanMngr      *ChannelManager
	// pluginMngr    *PluginManager
	// dictionary    *Dictionary
	database      *sqlx.DB
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

	// create the bot
	bot := Kabukibot{}
	bot.configuration = config
	bot.workers = make(map[string]*channelWorker)
	bot.channelMutex = sync.Mutex{}
	bot.logger = logger
	bot.twitch = twitch
	// bot.pluginMngr = NewPluginManager(&bot, dispatcher, db)
	// bot.dictionary = NewDictionary(db, logger)
	bot.database = nil
	bot.alive = make(chan struct{})

	return &bot, nil
}

func (bot *Kabukibot) Connect() error {
	// connect to database
	db, err := sqlx.Connect("mysql", bot.configuration.Database.DSN)
	if err != nil {
		return err
	}

	bot.database = db

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
	go bot.joinInitialChannels()

	prefix := bot.configuration.CommandPrefix

	for msg := range bot.twitch.Incoming() {
		// find the appropriate worker
		channel := msg.ChannelName()

		bot.channelMutex.Lock()
		worker, exists := bot.workers[channel]
		bot.channelMutex.Unlock()

		if exists {
			asserted, okay := msg.(twitch.TextMessage)
			if okay {
				worker.Input() <- TextMessage{asserted, prefix}
			} else {
				worker.Input() <- msg
			}
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

func (bot *Kabukibot) Database() *sqlx.DB {
	return bot.database
}

func (bot *Kabukibot) Logger() Logger {
	return bot.logger
}

// func (bot *Kabukibot) EmoteManager() EmoteManager {
// 	return bot.emoteMngr
// }

// func (bot *Kabukibot) Dictionary() *Dictionary {
// 	return bot.dictionary
// }

func (bot *Kabukibot) Channels() []string {
	bot.channelMutex.Lock()

	result := make([]string, 0)

	for cn, _ := range bot.workers {
		result = append(result, cn)
	}

	bot.channelMutex.Unlock()

	return result
}

// func (bot *Kabukibot) Channel(name string) (c *twitch.Channel, ok bool) {
// 	return bot.chanMngr.Channel(name)
// }

func (bot *Kabukibot) AddPlugin(plugin Plugin) {
	plugin.Setup(bot)

	bot.plugins = append(bot.plugins, plugin)
}

func (bot *Kabukibot) Plugins() []Plugin {
	return bot.plugins
}

func (bot *Kabukibot) Join(channel string) <-chan bool {
	channel = strings.ToLower(channel)

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
	worker := newChannelWorker(channel, bot)

	// remember worker
	bot.workers[channel] = worker
	bot.channelMutex.Unlock()

	// remember that we joined
	bot.Database().Exec("INSERT INTO channel (name) VALUES (?)", channel)

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
	channel = strings.ToLower(channel)

	// never leave our home channel
	if channel == "#"+strings.ToLower(bot.BotUsername()) {
		dummy := make(chan bool, 1)
		dummy <- false
		close(dummy)

		return dummy
	}

	bot.Database().Exec("DELETE FROM channel WHERE name = ?", channel)

	// send off the request to leave the channel, but wait for its confirmation
	// to shutdown our worker; this signal therefore does not represent the
	// moment when we left, but only the moment the PART message was sent.
	return bot.twitch.Send(twitch.PartMessage{channel})
}

func (bot *Kabukibot) Joined(channel string) bool {
	bot.channelMutex.Lock()
	_, joined := bot.workers[channel]
	bot.channelMutex.Unlock()

	return joined
}

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

type initialChannel struct {
	Name string `db:"name"`
}

func (bot *Kabukibot) joinInitialChannels() {
	// join the bot's channel
	bot.Join("#" + bot.BotUsername())

	// find previously joined channels
	list := make([]initialChannel, 0)
	db := bot.Database()

	db.Select(&list, "SELECT name FROM channel ORDER BY name")

	for _, channel := range list {
		bot.Join(channel.Name)
	}
}
