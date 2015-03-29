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
	bot, err := bot.NewKabukibot(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// add plugins
	bot.AddPlugin(plugin.NewCorePlugin())

	// here we go
	quit, err := bot.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// wait for disconnect
	<-quit
}
