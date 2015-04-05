package twitch

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
	mods  []string
}

func NewChannel(name string) *Channel {
	return &Channel{strings.ToLower(strings.TrimLeft(name, "#")), ChannelState{
		false,
		false,
		false,
		false,
		make([]int, 0),
	}, make([]string, 0)}
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

	if len(s.EmoteSet) > 0 {
		s.EmoteSet = make([]int, 0)
	}
}

func (c *Channel) findModerator(user string) int {
	for idx, u := range c.mods {
		if user == u {
			return idx
		}
	}

	return -1
}

func (c *Channel) IsModerator(user string) bool {
	return c.findModerator(user) != -1
}

// Adds a moderator to the channel.
// Returns true if the user was not yet added before, else false.
func (c *Channel) AddModerator(user string) bool {
	pos := c.findModerator(user)

	if pos == -1 {
		c.mods = append(c.mods, user)
		return true
	}

	return false
}

func (c *Channel) RemoveModerator(user string) bool {
	pos := c.findModerator(user)

	if pos != -1 {
		c.mods = append(c.mods[:pos], c.mods[(pos + 1):]...)
		return true
	}

	return false
}
