package twitch

import (
	"time"
	"strings"
	"strconv"
	irc "github.com/fluffle/goirc/client"
)

type TwitchClient struct {
	conn         *irc.Conn
	queue        SendQueue
	dispatcher   Dispatcher
	channels     map[string]*Channel
	ReadySignal  chan bool
	QuitSignal   chan bool
}

func NewTwitchClient(conn *irc.Conn, d Dispatcher, delay time.Duration) *TwitchClient {
	client := TwitchClient{
		conn,
		NewSendQueue(delay),
		d,
		make(map[string]*Channel),
		make(chan bool, 1),
		make(chan bool, 1),
	}

	client.setupInternalHandlers()

	return &client
}

func (client *TwitchClient) setupInternalHandlers() {
	client.conn.HandleFunc(irc.REGISTER,     client.onConnect)
	client.conn.HandleFunc(irc.DISCONNECTED, client.onDisconnect)
	client.conn.HandleFunc(irc.PRIVMSG,      client.onLine)
	client.conn.HandleFunc(irc.ACTION,       client.onLine)
	client.conn.HandleFunc(irc.MODE,         client.onMode)
	client.conn.HandleFunc(irc.JOIN,         client.onJoin)
	client.conn.HandleFunc(irc.PART,         client.onPart)
}

func (client *TwitchClient) Channel(name string) (c *Channel, ok bool) {
	c, ok = client.channels[strings.TrimLeft(name, "#")]
	return
}

func (client *TwitchClient) Connect() (chan bool, error) {
	// start working on our outgoing queue
	go client.queue.Worker()

	err := client.conn.Connect()
	if err != nil {
		return nil, err
	}

	return client.QuitSignal, nil
}

func (client *TwitchClient) Join(channel *Channel) {
	_, ok := client.Channel(channel.Name)
	if !ok {
		client.channels[channel.Name] = channel

		client.queue.Push(func() {
			client.conn.Join(channel.IrcName())
		})
	}
}

func (client *TwitchClient) Part(channel *Channel) {
	_, ok := client.Channel(channel.Name)
	if ok {
		client.queue.Push(func () {
			client.conn.Part(channel.IrcName())
			delete(client.channels, channel.Name)
		})
	}
}

func (client *TwitchClient) Privmsg(target string, text string) {
	client.queue.Push(func() {
		client.conn.Privmsg(target, text)
	})
}

func (client *TwitchClient) onConnect(conn *irc.Conn, line *irc.Line) {
	conn.Raw("TWITCHCLIENT 3")
	client.ReadySignal <- true
}

func (client *TwitchClient) onDisconnect(conn *irc.Conn, line *irc.Line) {
	client.QuitSignal <- true
}

func (client *TwitchClient) onJoin(conn *irc.Conn, line *irc.Line) {
	channel, ok := client.Channel(line.Target())
	if ok {
		client.dispatcher.HandleJoin(channel)
	}
}

func (client *TwitchClient) onPart(conn *irc.Conn, line *irc.Line) {
	channel, ok := client.Channel(line.Target())
	if ok {
		client.dispatcher.HandlePart(channel)
	}
}

func (client *TwitchClient) onMode(conn *irc.Conn, line *irc.Line) {
	channel, ok := client.Channel(line.Target())
	if !ok {
		return
	}

	mode, username := line.Args[1], line.Args[2]

	if mode == "+o" {
		channel.AddModerator(username)
	} else if mode == "-o" {
		channel.RemoveModerator(username)
	}
}

func (client *TwitchClient) onLine(conn *irc.Conn, line *irc.Line) {
	channel, ok := client.Channel(line.Target())
	if !ok {
		return
	}

	baseMsg := message{
		channel:   channel,
		user:      NewUser(line.Nick, channel),
		text:      line.Text(),
		time:      line.Time,
		processed: false,
	}

	// internal Twitch stuff, both state information (subscriber, turbo, emoteset, ...)
	// as well as one-time things like timeouts
	if line.Nick == "jtv" {
		parts   := strings.SplitN(baseMsg.text, " ", 3)
		command := strings.ToLower(parts[0])

		msg := &twitchMessage{
			baseMsg,
			command,
			parts[1:],
		}

		// handle the internal user state
		updateChannelState(msg)

		// now the world may know about this message
		client.dispatcher.HandleTwitchMessage(msg)

		return
	}

	// subscriber notifications
	if line.Nick == "twitchnotify" {
		client.dispatcher.HandleTwitchMessage(&twitchMessage{
			baseMsg,
			"SUBSCRIBER",
			make([]string, 0),
		})

		return
	}

	// handle someone typing "/me likes this"
	if line.Cmd == "ACTION" {
		baseMsg.text = "/me " + baseMsg.text
	}

	// pull the state we collected since the last text message and apply it to this user
	updateUserState(&baseMsg)

	// now tell the world what we got
	client.dispatcher.HandleTextMessage(&baseMsg)
}

func updateChannelState(msg *twitchMessage) {
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

func updateUserState(msg *message) {
	user  := msg.User()
	cn    := msg.Channel()
	state := &cn.State

	user.IsBroadcaster = user.Name == cn.Name
	user.IsModerator   = cn.IsModerator(user.Name)
	user.IsSubscriber  = state.Subscriber
	user.IsTurbo       = state.Turbo
	user.IsTwitchAdmin = state.Admin
	user.IsTwitchStaff = state.Staff
	user.EmoteSet      = state.EmoteSet

	state.Clear()
}
