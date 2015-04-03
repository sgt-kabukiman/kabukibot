package bot

type User struct {
	Name          string
	Channel       *Channel
	IsBot         bool
	IsOperator    bool
	IsBroadcaster bool
	IsModerator   bool
	IsSubscriber  bool
	IsTurbo       bool
	IsTwitchAdmin bool
	IsTwitchStaff bool
	EmoteSet      []int
}

func NewUser(name string, cnl *Channel) *User {
	return &User{Name: name, Channel: cnl}
}

func getChar(flag bool, sign string) string {
	if (flag) {
		return sign
	}

	return ""
}

func (u *User) Prefix() string {
	prefix := ""

	if u.IsBot           { prefix += "%" }
	if u.IsOperator      { prefix += "$" }
	if u.IsBroadcaster   { prefix += "&" }
	if u.IsModerator     { prefix += "@" }
	if u.IsSubscriber    { prefix += "+" }
	if u.IsTurbo         { prefix += "~" }
	if u.IsTwitchAdmin   { prefix += "!" }
	if u.IsTwitchStaff   { prefix += "!" }

	return prefix
}
