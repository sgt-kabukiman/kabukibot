package bot

import "time"

// interfaces

type Plugin interface {
	Setup(*Kabukibot, *Dispatcher)
}

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

type ModeMessage interface {
	Message

	Mode()    string
	Subject() string
}

type CommandMessage interface {
	Message

	Command() string
	Args()    []string
}

type Response interface {
	Channel() *Channel
	Text()    string
}

// structs

type MessageHandlerFunc func(Message)
type TextHandlerFunc func(TextMessage)
type TwitchHandlerFunc func(TwitchMessage)
type ModeHandlerFunc func(ModeMessage)
type CommandHandlerFunc func(CommandMessage)
type ProcessedHandlerFunc func(Message)
type ResponseHandlerFunc func(Response)

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

type modeMessage struct {
	message

	mode    string
	subject string
}

type commandMessage struct {
	message

	command string
	args    []string
}

type response struct {
	channel *Channel
	text    string
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

// CommandMessage interface
func (m *commandMessage) Command() string   { return m.command }
func (m *commandMessage) Args()    []string { return m.args    }

// ModeMessage interface
func (m *modeMessage) Mode()    string { return m.mode    }
func (m *modeMessage) Subject() string { return m.subject }

// Response interface
func (r *response) Channel() *Channel { return r.channel }
func (r *response) Text()    string   { return r.text    }
