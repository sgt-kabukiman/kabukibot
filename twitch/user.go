package twitch

type User struct {
	Name          string
	Channel       *Channel
	IsBroadcaster bool
	IsModerator   bool
	IsSubscriber  bool
	IsTurbo       bool
	IsTwitchAdmin bool
	IsTwitchStaff bool
	EmoteSet      []int
}

func NewUser(name string, cn *Channel) *User {
	return &User{Name: name, Channel: cn}
}
