package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
	kabukibot.AddPlugin(plugin.NewLogPlugin())
	kabukibot.AddPlugin(plugin.NewPingPlugin())
	kabukibot.AddPlugin(plugin.NewJoinPlugin())
	kabukibot.AddPlugin(plugin.NewACLPlugin())
	kabukibot.AddPlugin(plugin.NewPluginControlPlugin())
	// kabukibot.AddPlugin(plugin.NewSpeedrunComPlugin())
	kabukibot.AddPlugin(plugin.NewEchoPlugin())
	kabukibot.AddPlugin(plugin.NewSysInfoPlugin())
	kabukibot.AddPlugin(plugin.NewDictionaryPlugin())
	// kabukibot.AddPlugin(plugin.NewBanhammerBotPlugin())
	// kabukibot.AddPlugin(plugin.NewEmoteCounterPlugin())
	kabukibot.AddPlugin(plugin.NewSubHypePlugin())

	// here we go
	err = kabukibot.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// do your thing, kabukibot
	go kabukibot.Work()

	data, _ := ioutil.ReadFile("channels.txt")
	lines := strings.Split(string(data), "\n")

	for _, cn := range lines {
		kabukibot.Join(cn)
	}

	// wait for disconnect
	<-kabukibot.Alive()
}
