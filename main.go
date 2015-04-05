package main

import (
	"fmt"
	"os"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/twitch"
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
	kabukibot.AddPlugin(plugin.NewCorePlugin())
	kabukibot.AddPlugin(plugin.NewEchoPlugin())

	// here we go
	quit, err := kabukibot.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	kabukibot.Join(twitch.NewChannel("kabukibot"))
	kabukibot.Join(twitch.NewChannel("kabukibotdev"))

	// wait for disconnect
	<-quit
}
