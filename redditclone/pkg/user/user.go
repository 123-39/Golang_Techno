package user

type User struct {
	ID       string `json:"id"`
	Login    string `json:"username"`
	password string
}

type UserRepo interface {
	Authorize(login, pass string) (User, error)
	AddUser(login, pass string) (User, error)
}
