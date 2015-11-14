//go:generate stringer -type=FlagState,UserType -output=types_strings.go

package twitch

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/sorcix/irc"
)

type IncomingMessage interface {
	ChannelName() string
}

type OutgoingMessage interface {
	IrcMessage() *irc.Message
}

type RawMessage struct {
	Message irc.Message
}

func (self RawMessage) IrcMessage() *irc.Message {
	return &self.Message
}

type JoinMessage struct {
	Channel string
}

func (self JoinMessage) ChannelName() string {
	return self.Channel
}

func (self JoinMessage) IrcMessage() *irc.Message {
	return &irc.Message{
		Command: irc.JOIN,
		Params:  []string{self.Channel},
	}
}

type PartMessage struct {
	Channel string
}

func (self PartMessage) ChannelName() string {
	return self.Channel
}

func (self PartMessage) IrcMessage() *irc.Message {
	return &irc.Message{
		Command: irc.PART,
		Params:  []string{self.Channel},
	}
}

type RoomStateMessage struct {
	Channel  string
	IsNotice bool
	R9K      FlagState
	SlowMode FlagState
	SubsOnly FlagState
}

func (self RoomStateMessage) ChannelName() string {
	return self.Channel
}

type TextMessage struct {
	Channel string
	User    User
	Text    string
	Action  string
}

func (self TextMessage) ChannelName() string {
	return self.Channel
}

func (self TextMessage) IrcMessage() *irc.Message {
	return &irc.Message{
		Command:  irc.PRIVMSG,
		Params:   []string{self.Channel},
		Trailing: self.Text,
	}
}

type ClearChatMessage struct {
	Channel  string
	User     string
	Duration int // TODO: use time.Duration
}

func (self ClearChatMessage) ChannelName() string {
	return self.Channel
}

func (self ClearChatMessage) IrcMessage() *irc.Message {
	text := ""

	if self.User == "" {
		text = ".clearchat"
	} else {
		text = ".timeout " + self.User

		if self.Duration > 0 {
			text += " " + strconv.Itoa(self.Duration)
		}
	}

	return &irc.Message{
		Command:  irc.PRIVMSG,
		Params:   []string{self.Channel},
		Trailing: text,
	}
}

type SubscriberNotificationMessage struct {
	Channel string
	User    string
	Months  int
	Text    string
}

func (self SubscriberNotificationMessage) ChannelName() string {
	return self.Channel
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

type pongMessage struct {
	Params   []string
	Trailing string
}

func (self pongMessage) IrcMessage() *irc.Message {
	return &irc.Message{
		Command:  irc.PONG,
		Params:   self.Params,
		Trailing: self.Trailing,
	}
}

type capReqMessage struct {
	Capability string
}

func (self capReqMessage) IrcMessage() *irc.Message {
	return &irc.Message{
		Command: irc.CAP,
		Params:  []string{irc.CAP_REQ, ":twitch.tv/" + self.Capability},
	}
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
