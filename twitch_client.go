package main

import (
	// "time"
	// "fmt"
	"strings"
	"regexp"
	irc "github.com/fluffle/goirc/client"
)

type TwitchClient struct {
	conn       *irc.Conn
	dispatcher *Dispatcher
	ready      chan bool
	quit       chan bool
}

func NewTwitchClient(conn *irc.Conn, d *Dispatcher) *TwitchClient {
	client := TwitchClient{}
	client.conn       = conn
	client.dispatcher = d
	client.ready      = make(chan bool, 1)
	client.quit       = make(chan bool, 1)

	client.setupInternalHandlers()

	return &client
}

func (client *TwitchClient) setupInternalHandlers() {
	client.conn.HandleFunc(irc.REGISTER,     client.onConnect)
	client.conn.HandleFunc(irc.DISCONNECTED, client.onDisconnect)
	client.conn.HandleFunc(irc.PRIVMSG,      client.onLine)
	client.conn.HandleFunc(irc.MODE,         client.onLine)
}

func (client *TwitchClient) Connect() (chan bool, error) {
	err := client.conn.Connect()
	if err != nil {
		return nil, err
	}

	return client.quit, nil
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
		user:      line.Nick,
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
	client.dispatcher.HandleCommand(&cmd)
}
