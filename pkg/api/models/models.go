package models

type Test struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Score       uint64 `json:"score"`
}

type Stage struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Test   string `json:"test,omitempty"`
	Output string `json:"output,omitempty"`
}

func (s *Stage) Success() bool {
	return s.Status == "success"
}

type Commit struct {
	Login  string   `json:"login"`
	Repo   string   `json:"repository"`
	Commit string   `json:"commit"`
	Status string   `json:"status"`
	Checks []*Stage `json:"checks"`
}

type Run struct {
	Hash     string `json:"hash"`
	Status   string `json:"status"`
	Output   string `json:"output"`
	Score    uint64 `json:"score"`
	Test     string `json:"test"`
	Baseline bool   `json:"baseline"`
}

type Task struct {
	Id     string `json:"id"`
	Ref    string `json:"ref"`
	Url    string `json:"archive"`
	Stages []*Stage
	Runs   []*Run
}

type UserStat struct {
	Login string `json:"login"`
	Score uint   `json:"score"`
	Count uint   `json:"count"`
}

type UserEvent []*struct {
	Name    string `json:"name"`
	Updated string `json:"updated_at"`
	Status  string `json:"status"`
	Perf    uint   `json:"perf"`
}
