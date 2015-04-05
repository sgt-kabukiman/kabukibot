package plugin

import "fmt"
import "strings"
import "strconv"
import "github.com/sgt-kabukiman/kabukibot/bot"
import "github.com/sgt-kabukiman/kabukibot/twitch"

type CorePlugin struct {
	bot    *bot.Kabukibot
	config *bot.Configuration
	prefix string
}

func NewCorePlugin() *CorePlugin {
	return &CorePlugin{}
}

func (plugin *CorePlugin) Setup(bot *bot.Kabukibot, d *twitch.Dispatcher) {
	plugin.bot    = bot
	plugin.config = bot.Configuration()

	d.OnTextMessage(plugin.onText)
	d.OnTextMessage(plugin.printLine)
	d.OnTwitchMessage(plugin.onTwitch)
}

func (plugin *CorePlugin) onText(msg twitch.TextMessage) {
	user  := msg.User()
	cn    := msg.Channel()
	state := cn.State

	user.IsBot         = plugin.config.Account.Username == user.Name
	user.IsOperator    = plugin.config.Operator == user.Name
	user.IsBroadcaster = user.Name == cn.Name
	user.IsModerator   = cn.IsModerator(user.Name)
	user.IsSubscriber  = state.Subscriber
	user.IsTurbo       = state.Turbo
	user.IsTwitchAdmin = state.Admin
	user.IsTwitchStaff = state.Staff
	user.EmoteSet      = state.EmoteSet

	state.Clear()
}

func (plugin *CorePlugin) onTwitch(msg twitch.TwitchMessage) {
	cn := msg.Channel()

	switch msg.Command() {
	case "specialuser":
		args := msg.Args()

		switch args[1] {
		case "subscriber":
			cn.State.Subscriber = true
		case "turbo":
			cn.State.Turbo = true
		case "staff":
			cn.State.Staff = true
		case "admin":
			cn.State.Admin = true
		}

	case "emoteset":
		args := msg.Args()
		list := args[1]

		// trim "[" and "]"
		list = list[1:len(list)-1]

		codes := strings.Split(list, ",")
		ids   := make([]int, len(codes))

		for idx, code := range codes {
			converted, err := strconv.Atoi(code)
			if err == nil {
				ids[idx] = converted
			}
		}

		cn.State.EmoteSet = ids
	}
}

func (plugin* CorePlugin) printLine(msg twitch.TextMessage) {
	fmt.Printf("[#%v] %v%v: %v\n", msg.Channel().Name, msg.User().Prefix(), msg.User().Name, msg.Text())
}
