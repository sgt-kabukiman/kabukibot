//go:generate stringer -type=FlagState,UserType -output=types_strings.go
package twitch

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/sorcix/irc"
)

type Message interface {
}

type JoinMessage struct {
	Channel string
}

type PartMessage struct {
	Channel string
}

type RoomStateMessage struct {
	Channel  string
	IsNotice bool
	R9K      FlagState
	SlowMode FlagState
	SubsOnly FlagState
}

type TextMessage struct {
	Channel string
	User    User
	Text    string
	Action  string
}

type ClearChatMessage struct {
	Channel string
	User    string
}

type SubscriberNotificationMessage struct {
	Channel string
	User    string
	Months  int
	Text    string
}

var justSubscribed = regexp.MustCompile(`^([a-zA-Z0-9_]+) just subscribed!$`)
var reSubscribe = regexp.MustCompile(`^([a-zA-Z0-9_]+) subscribed for ([0-9]+) months in a row!$`)

func parseSubNotification(msg *irc.Message) SubscriberNotificationMessage {
	out := SubscriberNotificationMessage{
		Channel: msg.Params[0],
		User:    "",
		Months:  0,
		Text:    msg.Trailing,
	}

	match := justSubscribed.FindStringSubmatch(msg.Trailing)
	if len(match) > 0 {
		out.User = match[1]
	} else {
		match = reSubscribe.FindStringSubmatch(msg.Trailing)
		if len(match) > 0 {
			out.User = match[1]

			months, err := strconv.Atoi(match[2])
			if err != nil {
				out.Months = months
			}
		}
	}

	return out
}

type EmoticonMarker struct {
	FirstChar int
	LastChar  int
}

type EmoticonMarkers map[int][]EmoticonMarker

type User struct {
	Name       string
	Subscriber bool
	Turbo      bool
	ID         int
	Color      string
	Emotes     EmoticonMarkers
	Type       UserType
}

// Parses emoticon marker tags
//
// encoded is a string like "34:67-70,100-103/14:56-61"
func parseEmotesTag(encoded string) EmoticonMarkers {
	parts := strings.Split(encoded, "/")
	result := make(EmoticonMarkers)

	for _, part := range parts {
		subParts := strings.SplitN(part, ":", 2)

		emoteID, err := strconv.Atoi(subParts[0])
		if err != nil {
			continue
		}

		list := strings.Split(subParts[1], ",")
		positions := make([]EmoticonMarker, 0)

		for _, item := range list {
			itemParts := strings.SplitN(item, "-", 2)

			from, err := strconv.Atoi(itemParts[0])
			if err != nil {
				continue
			}

			to, err := strconv.Atoi(itemParts[1])
			if err != nil {
				continue
			}

			positions = append(positions, EmoticonMarker{from, to})
		}

		result[emoteID] = positions
	}

	return result
}

// type Message interface {
// 	Channel() *Channel
// 	User() *User
// 	Text() string
// 	Time() time.Time
// 	Processed() bool
// 	SetProcessed(bool)
// }

// type TextMessage interface {
// 	Message
// }

// type TwitchMessage interface {
// 	Message

// 	Command() string
// 	Args() []string
// }

// // structs

// type TextHandlerFunc func(TextMessage)
// type TwitchHandlerFunc func(TwitchMessage)
// type JoinHandlerFunc func(*Channel)

// type message struct {
// 	channel   *Channel
// 	user      *User
// 	text      string
// 	time      time.Time
// 	processed bool
// }

// type twitchMessage struct {
// 	message

// 	command string
// 	args    []string
// }

// // Message interface
// func (m *message) Channel() *Channel { return m.channel }
// func (m *message) User() *User       { return m.user }
// func (m *message) Text() string      { return m.text }
// func (m *message) Time() time.Time   { return m.time }
// func (m *message) Processed() bool   { return m.processed }

// func (m *message) SetProcessed(processed bool) {
// 	m.processed = processed
// }

// // TwitchMessage interface
// func (m *twitchMessage) Command() string { return m.command }
// func (m *twitchMessage) Args() []string  { return m.args }

type FlagState int

const (
	Enabled FlagState = iota
	Disabled
	Undefined
)

func parseFlagState(str string) FlagState {
	if str == "1" {
		return Enabled
	} else if str == "0" {
		return Disabled
	}

	return Undefined
}

type UserType int

const (
	Plebs UserType = iota
	Moderator
	GlobalModerator
	TwitchStaff
	TwitchAdmin
)

func parseUserType(t string) UserType {
	switch t {
	case "mod":
		return Moderator
	case "global_mod":
		return GlobalModerator
	case "staff":
		return TwitchStaff
	case "admin":
		return TwitchAdmin
	default:
		return Plebs
	}
}
