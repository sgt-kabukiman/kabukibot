package twitch

import "time"

type Message interface {
	Channel()   *Channel
	User()      *User
	Text()      string
	Time()      time.Time
	Processed() bool
}

type TextMessage interface {
	Message
}

type TwitchMessage interface {
	Message

	Command() string
	Args()    []string
}

// structs

type TextHandlerFunc func(TextMessage)
type TwitchHandlerFunc func(TwitchMessage)
type JoinHandlerFunc func(*Channel)

type message struct {
	channel   *Channel
	user      *User
	text      string
	time      time.Time
	processed bool
}

type twitchMessage struct {
	message

	command string
	args    []string
}

// Message interface
func (m *message) Channel()   *Channel  { return m.channel   }
func (m *message) User()      *User     { return m.user      }
func (m *message) Text()      string    { return m.text      }
func (m *message) Time()      time.Time { return m.time      }
func (m *message) Processed() bool      { return m.processed }

// TwitchMessage interface
func (m *twitchMessage) Command() string   { return m.command }
func (m *twitchMessage) Args()    []string { return m.args    }
