package main

import (
	"fmt"
	"os"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
)

func main() {
	// load configuration
	config, err := bot.LoadConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// build the bot
	kabukibot, err := bot.NewKabukibot(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// add plugins
	kabukibot.AddPlugin(plugin.NewBlacklistPlugin()) // load this as early as possible, because users will only be blacklisted for all following plugins
	kabukibot.AddPlugin(plugin.NewConsoleOutputPlugin())
	kabukibot.AddPlugin(plugin.NewPingPlugin())
	kabukibot.AddPlugin(plugin.NewJoinPlugin())
	kabukibot.AddPlugin(plugin.NewACLPlugin())
	kabukibot.AddPlugin(plugin.NewPluginControlPlugin())
	kabukibot.AddPlugin(plugin.NewSpeedrunComPlugin())
	kabukibot.AddPlugin(plugin.NewEchoPlugin())
	kabukibot.AddPlugin(plugin.NewSysInfoPlugin())
	kabukibot.AddPlugin(plugin.NewDictionaryControlPlugin())
	kabukibot.AddPlugin(plugin.NewBanhammerBotPlugin())
	kabukibot.AddPlugin(plugin.NewEmoteCounterPlugin())
	kabukibot.AddPlugin(plugin.NewSubHypePlugin())

	// here we go
	quit, err := kabukibot.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// wait for disconnect
	<-quit
}
