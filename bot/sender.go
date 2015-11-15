package bot

import (
	_ "github.com/go-sql-driver/mysql"

	"github.com/sgt-kabukiman/kabukibot/twitch"
)

type Sender interface {
	Send(twitch.OutgoingMessage) <-chan bool
	SendText(string) <-chan bool
}

// If ever neccessary, this can be tied to a channelWorker
// (e.g. if we were to have multiple IRC connections)
type senderStruct struct {
	twitch  *twitch.TwitchClient
	channel string
}

func newSenderStruct(client *twitch.TwitchClient, channel string) *senderStruct {
	return &senderStruct{client, channel}
}

func (self *senderStruct) Send(msg twitch.OutgoingMessage) <-chan bool {
	return self.twitch.Send(msg)
}

func (self *senderStruct) SendText(text string) <-chan bool {
	return self.Send(twitch.TextMessage{
		Channel: self.channel,
		Text:    text,
	})
}
