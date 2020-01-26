package github

type Repo struct {
	ID      int    `json:"id"`
	Owner   *User  `json:"owner"`
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

type User struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email,omitempty"`
}

type Installation struct {
	ID      int      `json:"id"`
	Account *Account `json:"account,omitempty"`
}

type Account struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

type CheckSuiteEvent struct {
	CheckSuite *CheckSuite   `json:"check_suite"`
	Repo       *Repo         `json:"repository"`
	Sender     *User         `json:"sender"`
	Inst       *Installation `json:"installation"`
}

type CheckSuite struct {
	Head string `json:"head_sha"`
}

type CheckRun struct {
	ID int `json:"id"`
}

type AccessToken struct {
	Token string `json:"token"`
}
