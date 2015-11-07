package plugin

import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type TestPlugin struct {
	channelPlugin

	bot    *bot.Kabukibot
	acl    *bot.ACL
	dict   *bot.Dictionary
	prefix string
}

func NewTestPlugin() *TestPlugin {
	return &TestPlugin{newChannelPlugin(), nil, nil, nil, ""}
}

func (plugin *TestPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	plugin.bot = bot
	plugin.acl = bot.ACL()
	plugin.dict = bot.Dictionary()
	plugin.prefix = bot.Configuration().CommandPrefix
}

func (plugin *TestPlugin) Key() string {
	return "testplugin"
}

func (plugin *TestPlugin) Permissions() []string {
	return []string{"test_perm1", "test_perm2"}
}

func (plugin *TestPlugin) Load(c *twitch.Channel, bot *bot.Kabukibot, d bot.Dispatcher) {
	plugin.addChannelListeners(c, listenerList{
		d.OnCommand(plugin.onCommand, c),
	})
}

func (plugin *TestPlugin) onCommand(cmd bot.Command) {
	command := cmd.Command()
	// channel := cmd.Channel()
	// acl     := plugin.acl

	println("       COMMAND in #" + cmd.Channel().Name + ": !" + command)

	// plugin.bot.Say(cmd.Channel(), plugin.dict.Get("crash_ctr_notext"))

	// switch command {
	// case "allow":
	// 	acl.Allow(channel.Name, cmd.Args()[0], "mypermission")
	// case "deny":
	// 	acl.Deny(channel.Name, cmd.Args()[0], "mypermission")
	// case "allowed":
	// 	if acl.IsAllowed(cmd.User(), "mypermission") {
	// 		println("allowed!")
	// 	} else {
	// 		println("not allowed")
	// 	}
	// }

	// println("done")

	// if (command == plugin.prefix + "say" || command == plugin.prefix + "echo") && plugin.bot.IsOperator(cmd.User().Name) {
	// 	plugin.bot.Say(cmd.Channel(), strings.Join(cmd.Args(), " "))
	// }
}
