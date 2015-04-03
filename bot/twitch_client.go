package bot

import (
	"time"
	// "fmt"
	"strings"
	"regexp"
	irc "github.com/fluffle/goirc/client"
)

type TwitchClient struct {
	conn       *irc.Conn
	queue      *SendQueue
	dispatcher *Dispatcher
	ready      chan bool
	quit       chan bool
}

func NewTwitchClient(conn *irc.Conn, d *Dispatcher, delay time.Duration) *TwitchClient {
	client := TwitchClient{conn, NewSendQueue(delay), d, make(chan bool, 1), make(chan bool, 1)}
	client.setupInternalHandlers()

	return &client
}

func (client *TwitchClient) setupInternalHandlers() {
	client.conn.HandleFunc(irc.REGISTER,     client.onConnect)
	client.conn.HandleFunc(irc.DISCONNECTED, client.onDisconnect)
	client.conn.HandleFunc(irc.PRIVMSG,      client.onLine)
	client.conn.HandleFunc(irc.MODE,         client.onLine)
	client.conn.HandleFunc(irc.ACTION,       client.onLine)
}

func (client *TwitchClient) Connect() (chan bool, error) {
	err := client.conn.Connect()
	if err != nil {
		return nil, err
	}

	return client.quit, nil
}

func (client *TwitchClient) Join(channel string) {
	client.queue.Push(func() { client.conn.Join(channel) })
}

func (client *TwitchClient) Part(channel string) {
	client.queue.Push(func () { client.conn.Part(channel) })
}

func (client *TwitchClient) onConnect(conn *irc.Conn, line *irc.Line) {
	conn.Raw("TWITCHCLIENT 3")
	client.ready <- true
}

func (client *TwitchClient) onDisconnect(conn *irc.Conn, line *irc.Line) {
	client.quit <- true
}

func (client *TwitchClient) onLine(conn *irc.Conn, line *irc.Line) {
	channel := NewChannel(line.Target())
	baseMsg := message{
		channel:   channel,
		user:      NewUser(line.Nick, channel),
		text:      line.Text(),
		time:      line.Time,
		processed: false,
	}

	// println("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
	// println("    Nick = " + line.Nick)
	// println("   Ident = " + line.Ident)
	// println("    Host = " + line.Host)
	// println("     Src = " + line.Src)
	// println("     Cmd = " + line.Cmd)
	// println("     Raw = " + line.Raw)
	// fmt.Printf("    Args = %v\n", line.Args)
	// println("Target() = " + line.Target())
	// println("  Text() = " + line.Text())
	// println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

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
	client.dispatcher.HandleMessage(msg)

	switch message := msg.(type) {
	case TwitchMessage:
		client.dispatcher.HandleTwitchMessage(message)
	case ModeMessage:
		client.dispatcher.HandleModeMessage(message)
	case Message:
		client.dispatcher.HandleTextMessage(message)
		client.processPossibleCommand(message)
	}

	client.dispatcher.HandleProcessed(msg)
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
	client.dispatcher.HandleCommandMessage(&cmd)
}
