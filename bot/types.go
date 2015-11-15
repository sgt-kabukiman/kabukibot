package bot

import (
	"regexp"
	"strings"

	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type TextMessage struct {
	twitch.TextMessage

	prefix   string
	operator string
}

func (self *TextMessage) IsCommand(cmd string) bool {
	return strings.HasPrefix(self.Text, "!"+cmd)
}

func (self *TextMessage) IsGlobalCommand(cmd string) bool {
	return self.IsCommand(self.prefix + cmd)
}

func (self *TextMessage) IsFrom(user string) bool {
	return strings.ToLower(self.User.Name) == strings.ToLower(user)
}

func (self *TextMessage) IsFromBroadcaster() bool {
	return self.IsFrom(strings.TrimPrefix(self.Channel, "#"))
}

func (self *TextMessage) IsFromOperator() bool {
	return self.IsFrom(self.operator)
}

func (self *TextMessage) IsFromBot() bool {
	return self.User.Myself
}

var commandRegex = regexp.MustCompile(`^!([a-zA-Z0-9_-]+)(?:\s+(.*))?$`)
var argSplitter = regexp.MustCompile(`\s+`)

func (self *TextMessage) Command() string {
	match := commandRegex.FindStringSubmatch(self.Text)
	if len(match) == 0 {
		return ""
	}

	return strings.ToLower(match[1])
}

func (self *TextMessage) Arguments() []string {
	args := make([]string, 0)

	match := commandRegex.FindStringSubmatch(self.Text)
	if len(match) == 0 {
		return args
	}

	argString := strings.TrimSpace(match[2])

	if len(argString) > 0 {
		args = argSplitter.Split(argString, -1)
	}

	return args
}

// type Command interface {
// 	twitch.Message

// 	Command() string
// 	Args() []string
// }

// type CommandHandlerFunc func(Command)

// type command struct {
// 	twitch.Message

// 	cmd  string
// 	args []string
// }

// func (cmd *command) Command() string { return cmd.cmd }
// func (cmd *command) Args() []string  { return cmd.args }

// type Response interface {
// 	ResponseTo() twitch.Message
// 	Channel() *twitch.Channel
// 	Text() string
// }

// type ResponseHandlerFunc func(Response)

// type response struct {
// 	to      twitch.Message
// 	channel *twitch.Channel
// 	text    string
// }

// func (r *response) ResponseTo() twitch.Message { return r.to }
// func (r *response) Channel() *twitch.Channel   { return r.channel }
// func (r *response) Text() string               { return r.text }

// type Plugin interface {
// 	Setup(*Kabukibot, Dispatcher)
// }

// type GlobalPlugin interface {
// 	Plugin
// }

// type ChannelPlugin interface {
// 	Plugin

// 	Key() string
// 	Permissions() []string
// 	Load(*twitch.Channel, *Kabukibot, Dispatcher)
// 	Unload(*twitch.Channel, *Kabukibot, Dispatcher)
// }
