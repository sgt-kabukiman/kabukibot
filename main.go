package main

import (
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin/acl"
	"github.com/sgt-kabukiman/kabukibot/plugin/banhammer_bot"
	"github.com/sgt-kabukiman/kabukibot/plugin/blacklist"
	"github.com/sgt-kabukiman/kabukibot/plugin/content"
	"github.com/sgt-kabukiman/kabukibot/plugin/custom_commands"
	"github.com/sgt-kabukiman/kabukibot/plugin/dictionary"
	"github.com/sgt-kabukiman/kabukibot/plugin/domain_ban"
	"github.com/sgt-kabukiman/kabukibot/plugin/echo"
	"github.com/sgt-kabukiman/kabukibot/plugin/emote_counter"
	"github.com/sgt-kabukiman/kabukibot/plugin/join"
	"github.com/sgt-kabukiman/kabukibot/plugin/log"
	"github.com/sgt-kabukiman/kabukibot/plugin/monitor"
	"github.com/sgt-kabukiman/kabukibot/plugin/ping"
	"github.com/sgt-kabukiman/kabukibot/plugin/plugin_control"
	"github.com/sgt-kabukiman/kabukibot/plugin/speedruncom"
	"github.com/sgt-kabukiman/kabukibot/plugin/subhype"
	"github.com/sgt-kabukiman/kabukibot/plugin/sysinfo"
	"github.com/sgt-kabukiman/kabukibot/plugin/troll"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

var (
	configFile   = kingpin.Flag("config", "Path to the config file to use").Required().String()
	channelsFile = kingpin.Flag("channels", "Path to a file of channels to join after connecting").String()
	debug        = kingpin.Flag("debug", "Enable debug output").Bool()
)

func main() {
	kingpin.Parse()

	// create logger
	level := bot.LogLevelInfo
	if *debug {
		level = bot.LogLevelDebug
	}

	logger := bot.NewLogger(level)

	// load configuration
	logger.Info("Loading configuration file @ " + *configFile + "...")
	config, err := bot.LoadConfiguration(*configFile)
	if err != nil {
		logger.Fatal(err.Error())
	}

	var channels []string

	if *channelsFile != "" {
		logger.Info("Reading initial channel list @ " + *channelsFile + "...")
		data, err := ioutil.ReadFile(*channelsFile)
		if err != nil {
			logger.Fatal(err.Error())
		}

		channels = strings.Split(string(data), "\n")
	}

	// connect to database
	logger.Info("Connecting to database...")
	db, err := sqlx.Connect("mysql", config.Database.DSN)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// setup our TwitchClient
	server := net.JoinHostPort(config.IRC.Host, strconv.Itoa(config.IRC.Port))
	twitch := twitch.NewTwitchClient(server, config.Account.Username, config.Account.Password, 2*time.Second)

	// build the bot
	kabukibot, err := bot.NewKabukibot(twitch, logger, db, config)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// add plugins
	kabukibot.AddPlugin(blacklist.NewPlugin()) // load this as early as possible, because users will only be blacklisted for all following plugins
	kabukibot.AddPlugin(log.NewPlugin())
	kabukibot.AddPlugin(ping.NewPlugin())
	kabukibot.AddPlugin(join.NewPlugin())
	kabukibot.AddPlugin(acl.NewPlugin())
	kabukibot.AddPlugin(plugin_control.NewPlugin())
	kabukibot.AddPlugin(speedruncom.NewPlugin())
	kabukibot.AddPlugin(echo.NewPlugin())
	kabukibot.AddPlugin(sysinfo.NewPlugin())
	kabukibot.AddPlugin(dictionary.NewPlugin())
	kabukibot.AddPlugin(domain_ban.NewPlugin())
	kabukibot.AddPlugin(banhammer_bot.NewPlugin())
	kabukibot.AddPlugin(emote_counter.NewPlugin())
	kabukibot.AddPlugin(subhype.NewPlugin())
	kabukibot.AddPlugin(troll.NewPlugin())
	kabukibot.AddPlugin(monitor.NewPlugin())
	kabukibot.AddPlugin(custom_commands.NewPlugin())
	kabukibot.AddPlugin(content.NewGTAPlugin())
	kabukibot.AddPlugin(content.NewCrashPlugin())
	kabukibot.AddPlugin(content.NewChattyPlugin())
	kabukibot.AddPlugin(content.NewSDAPlugin())
	kabukibot.AddPlugin(content.NewESAPlugin())

	// here we go
	err = kabukibot.Connect()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// do your thing, kabukibot
	logger.Info("Letting the magic happen...")
	go kabukibot.Work()

	for _, cn := range channels {
		<-kabukibot.Join(cn)
	}

	// wait for disconnect
	<-kabukibot.Alive()
}
