package main

import (
	irc "github.com/fluffle/goirc/client"
)

type TwitchClient struct {
	conn *irc.Conn
	quit chan bool
}

func NewTwitchClient(conn *irc.Conn) *TwitchClient {
	client := TwitchClient{}
	client.conn = conn
	client.setupInternalHandlers()

	return &client
}

func (client *TwitchClient) setupInternalHandlers() {
	client.quit = make(chan bool) // And a signal on disconnect

	client.On(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		client.quit <- true
	})

	client.On(irc.REGISTER, func(conn *irc.Conn, line *irc.Line) {
		conn.Raw("TWITCHCLIENT 3")
	})
}

func (client *TwitchClient) Connect() (chan bool, error) {
	err := client.conn.Connect()
	if err != nil {
		return nil, err
	}

	return client.quit, nil
}

func (client *TwitchClient) On(event string, fn irc.HandlerFunc) irc.Remover {
	return client.conn.Handle(event, fn)
}

func (client *TwitchClient) OnBG(event string, fn irc.HandlerFunc) irc.Remover {
	return client.conn.HandleBG(event, fn)
}
