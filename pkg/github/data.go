package github

type Repo struct {
	ID      int    `json:"id"`
	Owner   *User  `json:"owner"`
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

type User struct {
	ID    uint64 `json:"id"`
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
	Action     string        `json:"action"`
	Repo       *Repo         `json:"repository"`
	Sender     *User         `json:"sender"`
	Inst       *Installation `json:"installation"`
}

type CheckSuite struct {
	Head string `json:"head_sha"`
}

type CheckRun struct {
	ID             uint64          `json:"id,omitempty"`
	Name           string          `json:"name,omitempty"`
	Url            string          `json:"details_url,omitempty"`
	Commit         string          `json:"head_sha,omitempty"`
	Status         string          `json:"status,omitempty"`
	Conclusion     string          `json:"conclusion,omitempty"`
	StartTime      string          `json:"started_at,omitempty"`
	CompletionTime string          `json:"completed_at,omitempty"`
	Output         *CheckRunOutput `json:"output,omitempty"`
}

type CheckRunOutput struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type AccessToken struct {
	Token string `json:"token"`
}
