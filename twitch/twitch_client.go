package twitch

import (
	"time"
	"strings"
	"regexp"
	irc "github.com/fluffle/goirc/client"
)

type TwitchClient struct {
	conn         *irc.Conn
	queue        SendQueue
	dispatcher   *Dispatcher
	channels     map[string]*Channel
	ReadySignal  chan bool
	QuitSignal   chan bool
}

func NewTwitchClient(conn *irc.Conn, d *Dispatcher, delay time.Duration) *TwitchClient {
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
	client.conn.HandleFunc(irc.MODE,         client.onLine)
	client.conn.HandleFunc(irc.ACTION,       client.onLine)
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
		client.dispatcher.handleJoin(channel)
	}
}

func (client *TwitchClient) onPart(conn *irc.Conn, line *irc.Line) {
	channel, ok := client.Channel(line.Target())
	if ok {
		client.dispatcher.handlePart(channel)
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

	if line.Cmd == "MODE" {
		client.handleMessage(&modeMessage{
			baseMsg,
			line.Args[1],
			line.Args[2],
		})
	} else if line.Cmd == "ACTION" {
		baseMsg.text = "/me " + baseMsg.text
		client.handleMessage(&baseMsg)
	} else if line.Nick == "jtv" {
		parts   := strings.SplitN(baseMsg.text, " ", 3)
		command := strings.ToLower(parts[0])

		client.handleMessage(&twitchMessage{
			baseMsg,
			command,
			parts[1:],
		})
	} else if line.Nick == "twitchnotify" {
		client.handleMessage(&twitchMessage{
			baseMsg,
			"SUBSCRIBER",
			make([]string, 0),
		})
	} else {
		client.handleMessage(&baseMsg)
	}
}

func (client *TwitchClient) handleMessage(msg Message) {
	client.dispatcher.handleMessage(msg)

	switch message := msg.(type) {
	case TwitchMessage:
		client.dispatcher.handleTwitchMessage(message)
	case ModeMessage:
		client.dispatcher.handleModeMessage(message)
	case Message:
		client.dispatcher.handleTextMessage(message)
		client.processPossibleCommand(message)
	}

	client.dispatcher.handleProcessed(msg)
}

var commandRegex = regexp.MustCompile(`^!([a-zA-Z0-9_-]+)(?:\s+(.*))?$`)
var argSplitter  = regexp.MustCompile(`\s+`)

func (client *TwitchClient) processPossibleCommand(msg Message) {
	match := commandRegex.FindStringSubmatch(msg.Text())
	if len(match) == 0 {
		return
	}

	baseMsg   := msg.(*message)
	command   := strings.ToLower(match[1])
	argString := strings.TrimSpace(match[2])
	args      := make([]string, 0)

	if len(argString) > 0 {
		args = argSplitter.Split(argString, -1)
	}

	cmd := commandMessage{*baseMsg, command, args}
	client.dispatcher.handleCommandMessage(&cmd)
}
