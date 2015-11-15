package bot

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type Sender interface {
	Send(twitch.OutgoingMessage) <-chan bool
	SendText(string) <-chan bool
	Respond(string) <-chan bool
}

// If ever neccessary, this can be tied to a channelWorker
// (e.g. if we were to have multiple IRC connections)
type channelSender struct {
	twitch  *twitch.TwitchClient
	channel string
}

func newChannelSender(client *twitch.TwitchClient, channel string) *channelSender {
	return &channelSender{client, channel}
}

func (self *channelSender) newResponder(msg *TextMessage) *responder {
	return &responder{self, msg}
}

func (self *channelSender) Send(msg twitch.OutgoingMessage) <-chan bool {
	return self.twitch.Send(msg)
}

func (self *channelSender) SendText(text string) <-chan bool {
	return self.Send(twitch.TextMessage{
		Channel: self.channel,
		Text:    text,
	})
}

func (self *channelSender) Respond(text string) <-chan bool {
	return self.SendText(text)
}

// a sender that is tied to a received message and can be used to transparently address the
// original sender by name
type responder struct {
	cn  *channelSender
	msg *TextMessage
}

func (self *responder) Send(msg twitch.OutgoingMessage) <-chan bool {
	return self.cn.Send(msg)
}

func (self *responder) SendText(text string) <-chan bool {
	return self.cn.SendText(text)
}

func (self *responder) Respond(text string) <-chan bool {
	return self.SendText(fmt.Sprintf("%s, %s", self.msg.User.Name, text))
}
