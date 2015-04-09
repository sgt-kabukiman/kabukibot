package plugin

import "fmt"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type ConsoleOutputPlugin struct {
	bot *bot.Kabukibot
	me  string
}

func NewConsoleOutputPlugin() *ConsoleOutputPlugin {
	return &ConsoleOutputPlugin{}
}

func (plugin *ConsoleOutputPlugin) Setup(bot *bot.Kabukibot, d bot.Dispatcher) {
	plugin.bot = bot
	plugin.me  = bot.Configuration().Account.Username

	d.OnTextMessage(plugin.printLine, nil)
	d.OnResponse(plugin.onResponse, nil)
}

func (plugin* ConsoleOutputPlugin) printLine(msg twitch.TextMessage) {
	user := msg.User()

	fmt.Printf("[#%v] %v%v: %v\n", msg.Channel().Name, userPrefix(user), user.Name, msg.Text())
}

func (plugin* ConsoleOutputPlugin) onResponse(r bot.Response) {
	fmt.Printf("[#%v] %%%v: %v\n", r.Channel().Name, plugin.me, r.Text())
}

func getChar(flag bool, sign string) string {
	if (flag) {
		return sign
	}

	return ""
}

func userPrefix(u *twitch.User) string {
	prefix := ""

	// if u.IsOperator      { prefix += "$" }
	if u.IsBroadcaster   { prefix += "&" }
	if u.IsModerator     { prefix += "@" }
	if u.IsSubscriber    { prefix += "+" }
	if u.IsTurbo         { prefix += "~" }
	if u.IsTwitchAdmin   { prefix += "!" }
	if u.IsTwitchStaff   { prefix += "!" }

	return prefix
}
