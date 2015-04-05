package twitch

type UserManager interface {
	IsOperator(username string) bool
}

type userManager struct {
	operator    string
	subscribers map[string][]string
}

func (u *userManager) IsOperator(username string) bool {
	return false
}
