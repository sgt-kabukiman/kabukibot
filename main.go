package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	"github.com/sgt-kabukiman/kabukibot/plugin/blacklist"
	"github.com/sgt-kabukiman/kabukibot/twitch"
)

func main() {
	// load configuration
	config, err := bot.LoadConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create logger
	logger := bot.NewLogger(bot.LOG_LEVEL_DEBUG)

	// setup our TwitchClient
	server := net.JoinHostPort(config.IRC.Host, strconv.Itoa(config.IRC.Port))
	twitch := twitch.NewTwitchClient(server, config.Account.Username, config.Account.Password, 2*time.Second)

	// build the bot
	kabukibot, err := bot.NewKabukibot(twitch, logger, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// add plugins
	kabukibot.AddPlugin(blacklist.NewPlugin()) // load this as early as possible, because users will only be blacklisted for all following plugins
	kabukibot.AddPlugin(plugin.NewLogPlugin())
	kabukibot.AddPlugin(plugin.NewPingPlugin())
	kabukibot.AddPlugin(plugin.NewJoinPlugin())
	kabukibot.AddPlugin(plugin.NewACLPlugin())
	kabukibot.AddPlugin(plugin.NewPluginControlPlugin())
	kabukibot.AddPlugin(plugin.NewSpeedrunComPlugin())
	kabukibot.AddPlugin(plugin.NewEchoPlugin())
	kabukibot.AddPlugin(plugin.NewSysInfoPlugin())
	kabukibot.AddPlugin(plugin.NewDictionaryPlugin())
	kabukibot.AddPlugin(plugin.NewDomainBanPlugin())
	kabukibot.AddPlugin(plugin.NewBanhammerBotPlugin())
	kabukibot.AddPlugin(plugin.NewEmoteCounterPlugin())
	kabukibot.AddPlugin(plugin.NewSubHypePlugin())
	kabukibot.AddPlugin(plugin.NewTrollPlugin())
	kabukibot.AddPlugin(plugin.NewMonitorPlugin())
	kabukibot.AddPlugin(plugin.NewCustomCommandsPlugin())

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
