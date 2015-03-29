package main

import "strings"

type Channel struct {
	Name string
}

func NewChannel(name string) *Channel {
	return &Channel{strings.TrimLeft(name, "#")}
}

func (c *Channel) IrcName() string {
	return "#" + c.Name
}
