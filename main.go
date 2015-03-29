package main

import (
	"fmt"
	"os"

	"github.com/sgt-kabukiman/kabukibot/plugin"
)

var bot *Kabukibot

func main() {
	m := TwitchMessage{}

	func(m Message) {
		fmt.Println(m)
	}(m)

	os.Exit(0)

	// load configuration
	config, err := LoadConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// build the bot
	bot, err := NewKabukibot(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// add plugins
	testPlugin := plugin.NewTestPlugin()
	bot.AddPlugin(&testPlugin)

	// here we go
	quit, err := bot.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// wait for disconnect
	<-quit
}
