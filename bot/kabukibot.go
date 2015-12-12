package bot

import (
	"errors"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type Kabukibot struct {
	twitch        twitch.Client
	workers       map[string]*channelWorker
	channelMutex  sync.Mutex
	plugins       []Plugin
	logger        Logger
	dictionary    *Dictionary
	database      *sqlx.DB
	configuration *Configuration
	alive         chan struct{}
}

func NewKabukibot(client twitch.Client, log Logger, db *sqlx.DB, config *Configuration) (*Kabukibot, error) {
	// create the bot
	bot := Kabukibot{}
	bot.database = db
	bot.configuration = config
	bot.workers = make(map[string]*channelWorker)
	bot.channelMutex = sync.Mutex{}
	bot.logger = log
	bot.twitch = client
	bot.alive = make(chan struct{})

	return &bot, nil
}

func (bot *Kabukibot) Connect() error {
	// load dictionary elements
	bot.logger.Debug("Loading dictionary...")
	bot.dictionary = NewDictionary(bot.database, bot.logger)
	bot.dictionary.load()

	// setup plugins
	bot.logger.Debug("Setting up plugins...")
	for _, plugin := range bot.plugins {
		plugin.Setup(bot)
	}

	// connect to Twitch
	client := bot.twitch

	bot.logger.Info("Connecting to Twitch chat @ %s:%d...", bot.configuration.IRC.Host, bot.configuration.IRC.Port)
	err := client.Connect()
	if err != nil {
		return err
	}

	// wait for the ready signal
	<-client.Ready()
	bot.logger.Info("Connection established.")

	return nil
}

func (bot *Kabukibot) Shutdown() {
	// shutdown all channel workers
	bot.channelMutex.Lock()

	bot.logger.Info("Beginning shutdown procedure...")

	wg := sync.WaitGroup{}
	wg.Add(len(bot.workers))

	for _, worker := range bot.workers {
		signal := worker.Shutdown()

		go func() {
			<-signal
			wg.Done()
		}()
	}

	wg.Wait()
	bot.channelMutex.Unlock()

	bot.logger.Info("All channel workers have shut down.")

	// disconnect from IRC;
	// This will close the twitch client's incoming channel and hence stop .Work(),
	// which will close self.alive eventually.
	bot.logger.Info("Disconnecting from IRC...")
	bot.twitch.Disconnect()

	<-bot.alive
	bot.logger.Info("It's dead, Jim.")
}

func (bot *Kabukibot) Work() {
	go bot.joinInitialChannels()

	prefix := bot.configuration.CommandPrefix
	operator := bot.OpUsername()

	for msg := range bot.twitch.Incoming() {
		// find the appropriate worker
		channel := msg.ChannelName()

		bot.channelMutex.Lock()
		worker, exists := bot.workers[channel]
		bot.channelMutex.Unlock()

		if exists {
			asserted, okay := msg.(twitch.TextMessage)
			if okay {
				worker.Input() <- TextMessage{asserted, prefix, operator, false}
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

func (bot *Kabukibot) Dictionary() *Dictionary {
	return bot.dictionary
}

func (bot *Kabukibot) Channel(name string) (Channel, error) {
	bot.channelMutex.Lock()
	defer bot.channelMutex.Unlock()

	for cn, worker := range bot.workers {
		if cn == name {
			return worker, nil
		}
	}

	return nil, errors.New("Channel not found")
}

func (bot *Kabukibot) Channels() []string {
	bot.channelMutex.Lock()

	result := make([]string, 0)

	for cn, _ := range bot.workers {
		result = append(result, cn)
	}

	bot.channelMutex.Unlock()

	return result
}

func (bot *Kabukibot) AddPlugin(plugin Plugin) {
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

	bot.logger.Info("Joining %s...", channel)

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

	bot.logger.Info("Leaving %s...", channel)

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

func (bot *Kabukibot) QueueLen() int {
	return bot.twitch.QueueLen()
}

func (bot *Kabukibot) MessagesSent() int {
	return bot.twitch.MessagesSent()
}

func (bot *Kabukibot) MessagesReceived() int {
	return bot.twitch.MessagesReceived()
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
		<-bot.Join(channel.Name)
	}
}
