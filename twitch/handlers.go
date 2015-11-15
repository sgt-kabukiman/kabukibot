package twitch

import (
	"log"
	"strconv"
	"strings"

	"github.com/sorcix/irc"
)

type HandlerFunc func(*irc.Message, irc.Tags)

func (client *TwitchClient) setupHandlers() {
	client.handlers = map[string]HandlerFunc{
		irc.RPL_WELCOME: client.onWelcome,
		irc.PING:        client.onPing,
		irc.JOIN:        client.onJoin,
		irc.PART:        client.onPart,
		irc.PRIVMSG:     client.onPrivmsg,

		// special twitch commands
		"ROOMSTATE": client.onRoomState,
		"NOTICE":    client.onRoomState, // re-use the handler
		"CLEARCHAT": client.onClearChat,
	}
}

func (client *TwitchClient) onWelcome(msg *irc.Message, tags irc.Tags) {
	var sent <-chan bool

	caps := []string{"membership", "commands", "tags"}

	for _, capability := range caps {
		sent = client.Send(capReqMessage{capability})
	}

	// wait for the message being sent
	okay := <-sent

	if !okay {
		log.Fatal("Could not sent capabilities. Cannot procede.")
	}

	// signal to the outside world that now everything is set up
	close(client.ready)
}

func (client *TwitchClient) onPing(msg *irc.Message, tags irc.Tags) {
	client.Send(pongMessage{msg.Params, msg.Trailing})
}

func (client *TwitchClient) onJoin(msg *irc.Message, tags irc.Tags) {
	// only care about when WE joined something
	if msg.Prefix != nil && msg.Prefix.User == client.username {
		client.incoming <- JoinMessage{msg.Params[0]}
	}
}

func (client *TwitchClient) onPart(msg *irc.Message, tags irc.Tags) {
	// only care about when WE parted something
	if msg.Prefix != nil && msg.Prefix.User == client.username {
		client.incoming <- PartMessage{msg.Params[0]}
	}
}

func (client *TwitchClient) onRoomState(msg *irc.Message, tags irc.Tags) {
	message := RoomStateMessage{Channel: msg.Params[0], IsNotice: false}

	if msg.Command == "NOTICE" {
		message.IsNotice = true
	}

	flag, _ := tags["subs-only"]
	message.SubsOnly = parseFlagState(flag)

	flag, _ = tags["slow"]
	message.SlowMode = parseFlagState(flag)

	flag, _ = tags["r9k"]
	message.R9K = parseFlagState(flag)

	client.incoming <- message
}

func (client *TwitchClient) onPrivmsg(msg *irc.Message, tags irc.Tags) {
	nickname := ""

	if msg.Prefix != nil {
		nickname = msg.Prefix.User
	}

	// distinguish between regular messages and subscriber notifications
	if nickname == "twitchnotify" {
		client.incoming <- parseSubNotification(msg)
		return
	}

	// parse user information from tags
	user := User{Name: nickname, Type: Plebs}

	displayName, okay := tags["display-name"]
	if okay && len(displayName) > 0 {
		user.Name = displayName
	}

	flag, okay := tags["subscriber"]
	if okay {
		user.Subscriber = (flag == "1")
	}

	flag, okay = tags["turbo"]
	if okay {
		user.Turbo = (flag == "1")
	}

	value, okay := tags["user-id"]
	if okay {
		id, err := strconv.Atoi(value)
		if err != nil {
			user.ID = id
		}
	}

	value, okay = tags["color"]
	if okay {
		user.Color = value
	}

	value, okay = tags["emotes"]
	if okay {
		user.Emotes = parseEmotesTag(value)
	}

	value, okay = tags["user-type"]
	if okay {
		user.Type = parseUserType(value)
	}

	// handle IRC ACTION commands
	text := msg.Trailing
	action := ""

	if strings.HasPrefix(text, "\x01") {
		text = strings.Trim(text, "\x01")

		if strings.HasPrefix(text, "ACTION ") {
			action = "/me"
			text = strings.TrimPrefix(text, "ACTION ")
		}
	}

	message := TextMessage{
		Channel: msg.Params[0],
		User:    user,
		Text:    text,
		Action:  action,
	}

	client.incoming <- message
}

func (client *TwitchClient) onClearChat(msg *irc.Message, tags irc.Tags) {
	client.incoming <- ClearChatMessage{
		Channel: msg.Params[0],
		User:    msg.Trailing,
	}
}
