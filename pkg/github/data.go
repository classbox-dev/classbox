package github

type Repo struct {
	ID       int    `json:"id"`
	Owner    *User  `json:"owner"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
}

type User struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

type Installation struct {
	ID      int      `json:"id"`
	Account *Account `json:"account"`
}

type Account struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}
