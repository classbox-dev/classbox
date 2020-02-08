package models

import (
	"fmt"
	"time"
)

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

func (t *Stage) FillFromRun(stageName string, run *Run) {
	t.Name = fmt.Sprintf("%s::%s", stageName, run.Test)
	t.Status = run.Status
	t.Test = run.Test
	t.Output = run.Output
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

func (r *Run) CompareToBaseline(b *Run) {
	if r.Status != "success" {
		return
	}
	percent := r.Score * 1000 / b.Score
	humanPercent := float64(percent) / 10.
	r.Output = fmt.Sprintf("%.1f%% of baseline", humanPercent)
	if percent > 1200 {
		r.Status = "failure"
	}
}

type Task struct {
	Id     string `json:"id"`
	Ref    string `json:"ref"`
	Url    string `json:"archive"`
	Stages []*Stage
	Runs   []*Run
}

func (t *Task) ReportSystemError(test string) {
	var name string
	if test == "" {
		name = "system"
	} else {
		name = fmt.Sprintf("test::%s", test)
	}
	t.Stages = append(t.Stages, &Stage{
		Name:   name,
		Status: "exception",
		Test:   test,
		Output: "System error. Reported to administrators.",
	})
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

type Course struct {
	Update time.Time `json:"updated_at,omitempty"`
	Ready  bool      `json:"is_ready"`
}