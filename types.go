package main

import (
	"time"

	// irc "github.com/fluffle/goirc/client"
)

type Plugin interface {

}

type Message struct {
	Channel   *Channel
	User      string
	Text      string
	Time      time.Time
	Processed bool
}

type TextMessage Message
type ModeMessage Message

type TwitchMessage struct {
	Message
}

type Response struct {
	Channel *Channel
	Text    string
}

type MessageHandlerFunc func(*Message)
type TextHandlerFunc func(*TextMessage)
type TwitchHandlerFunc func(*TwitchMessage)
type ModeHandlerFunc func(*ModeMessage)
type CommandHandlerFunc func(command string, args []string, msg *Message)
type ProcessedHandlerFunc func(*Message)
type ResponseHandlerFunc func(*Response)
