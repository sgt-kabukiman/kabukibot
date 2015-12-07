package main

import (
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
	"github.com/sgt-kabukiman/kabukibot/test"
)

func initTester(t *test.Tester) {
	t.AddPlugin("blacklist", func() bot.Plugin {
		return blacklist.NewPlugin()
	})

	t.AddPlugin("log", func() bot.Plugin {
		return log.NewPlugin()
	})

	t.AddPlugin("ping", func() bot.Plugin {
		return ping.NewPlugin()
	})

	t.AddPlugin("join", func() bot.Plugin {
		return join.NewPlugin()
	})

	t.AddPlugin("acl", func() bot.Plugin {
		return acl.NewPlugin()
	})

	t.AddPlugin("plugin_control", func() bot.Plugin {
		return plugin_control.NewPlugin()
	})

	t.AddPlugin("speedruncom", func() bot.Plugin {
		return speedruncom.NewPlugin()
	})

	t.AddPlugin("echo", func() bot.Plugin {
		return echo.NewPlugin()
	})

	t.AddPlugin("sysinfo", func() bot.Plugin {
		return sysinfo.NewPlugin()
	})

	t.AddPlugin("dictionary", func() bot.Plugin {
		return dictionary.NewPlugin()
	})

	t.AddPlugin("domain_ban", func() bot.Plugin {
		return domain_ban.NewPlugin()
	})

	t.AddPlugin("banhammer_bot", func() bot.Plugin {
		return banhammer_bot.NewPlugin()
	})

	t.AddPlugin("emote_counter", func() bot.Plugin {
		return emote_counter.NewPlugin()
	})

	t.AddPlugin("subhype", func() bot.Plugin {
		return subhype.NewPlugin()
	})

	t.AddPlugin("troll", func() bot.Plugin {
		return troll.NewPlugin()
	})

	t.AddPlugin("monitor", func() bot.Plugin {
		return monitor.NewPlugin()
	})

	t.AddPlugin("custom_commands", func() bot.Plugin {
		return custom_commands.NewPlugin()
	})
}
