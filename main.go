package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin/acl"
	"github.com/sgt-kabukiman/kabukibot/plugin/banhammer_bot"
	"github.com/sgt-kabukiman/kabukibot/plugin/blacklist"
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

func main() {
	// load configuration
	config, err := bot.LoadConfiguration("config.yaml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create logger
	logger := bot.NewLogger(bot.LOG_LEVEL_DEBUG)

	// setup our TwitchClient
	server := net.JoinHostPort(config.IRC.Host, strconv.Itoa(config.IRC.Port))
	twitch := twitch.NewTwitchClient(server, config.Account.Username, config.Account.Password, 2*time.Second)

	// connect to database
	db, err := sqlx.Connect("mysql", config.Database.DSN)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// build the bot
	kabukibot, err := bot.NewKabukibot(twitch, logger, db, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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

	// here we go
	err = kabukibot.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// do your thing, kabukibot
	go kabukibot.Work()

	// data, _ := ioutil.ReadFile("channels.txt")
	// lines := strings.Split(string(data), "\n")

	// for _, cn := range lines {
	// 	kabukibot.Join(cn)
	// }

	// wait for disconnect
	<-kabukibot.Alive()
}
