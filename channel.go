package main

type Channel struct {
	Name string
}

func (c *Channel) GetIrcName() string {
	return "#" + c.Name
}
