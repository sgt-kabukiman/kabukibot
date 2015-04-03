package bot

import "strings"

type ChannelState struct {
	Subscriber bool
	Turbo      bool
	Staff      bool
	Admin      bool
	EmoteSet   []int
}

type Channel struct {
	Name  string
	State ChannelState
}

func NewChannel(name string) *Channel {
	return &Channel{strings.TrimLeft(name, "#"), ChannelState{
		false,
		false,
		false,
		false,
		make([]int, 0),
	}}
}

func (c *Channel) IrcName() string {
	return "#" + c.Name
}

func (c *Channel) ClearState() {
	c.State.Clear()
}

func (s *ChannelState) Clear() {
	s.Subscriber = false
	s.Turbo      = false
	s.Staff      = false
	s.Admin      = false
	s.EmoteSet   = make([]int, 0)
}
