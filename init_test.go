package main

import (
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	"github.com/sgt-kabukiman/kabukibot/plugin/blacklist"
	"github.com/sgt-kabukiman/kabukibot/test"
)

func initTester(t *test.Tester) {
	t.AddPlugin("blacklist", func() bot.Plugin {
		return blacklist.NewPlugin()
	})

	t.AddPlugin("log", func() bot.Plugin {
		return plugin.NewLogPlugin()
	})

	t.AddPlugin("ping", func() bot.Plugin {
		return plugin.NewPingPlugin()
	})

	t.AddPlugin("join", func() bot.Plugin {
		return plugin.NewJoinPlugin()
	})

	t.AddPlugin("acl", func() bot.Plugin {
		return plugin.NewACLPlugin()
	})

	t.AddPlugin("plugin_control", func() bot.Plugin {
		return plugin.NewPluginControlPlugin()
	})

	t.AddPlugin("speedruncom", func() bot.Plugin {
		return plugin.NewSpeedrunComPlugin()
	})

	t.AddPlugin("echo", func() bot.Plugin {
		return plugin.NewEchoPlugin()
	})

	t.AddPlugin("sysinfo", func() bot.Plugin {
		return plugin.NewSysInfoPlugin()
	})

	t.AddPlugin("dictionary", func() bot.Plugin {
		return plugin.NewDictionaryPlugin()
	})

	t.AddPlugin("domain_ban", func() bot.Plugin {
		return plugin.NewDomainBanPlugin()
	})

	t.AddPlugin("banhammer_bot", func() bot.Plugin {
		return plugin.NewBanhammerBotPlugin()
	})

	t.AddPlugin("emote_counter", func() bot.Plugin {
		return plugin.NewEmoteCounterPlugin()
	})

	t.AddPlugin("subhype", func() bot.Plugin {
		return plugin.NewSubHypePlugin()
	})

	t.AddPlugin("troll", func() bot.Plugin {
		return plugin.NewTrollPlugin()
	})

	t.AddPlugin("monitor", func() bot.Plugin {
		return plugin.NewMonitorPlugin()
	})

	t.AddPlugin("custom_commands", func() bot.Plugin {
		return plugin.NewCustomCommandsPlugin()
	})
}
