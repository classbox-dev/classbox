package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type Stage struct {
	Status string `json:"status"`
	Name   string `json:"name"`
	Test   string `json:"test,omitempty"`
	Output string `json:"output,omitempty"`
}

func (s *Stage) Success() bool {
	return s.Status == "success"
}

type Result struct {
	ExitCode int
	Output   []byte
	Stages   []Stage
}

func (r *Result) Success() bool {
	return r.ExitCode == 0
}

func RunStaged(ctx context.Context, volumes map[string]string, args ...string) (*Result, error) {
	r, err := Run(ctx, volumes, args...)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(r.Output, &r.Stages)
	if err != nil {
		return nil, fmt.Errorf("run error: %v", string(r.Output))
	}
	return r, nil
}

func Run(ctx context.Context, volumes map[string]string, args ...string) (*Result, error) {
	cArgs := []string{"run", "--rm"}
	for s, t := range volumes {
		cArgs = append(cArgs, "-v", fmt.Sprintf("%s:%s", s, t))
	}
	cArgs = append(cArgs, args...)

	cmd := exec.CommandContext(ctx, "docker", cArgs...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	st := Result{}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			st.ExitCode = exitError.ExitCode()
		} else {
			return nil, err
		}
	}
	st.Output = out.Bytes()
	return &st, nil
}

func BuildTests(ctx context.Context, url string) *Result {
	r, err := RunStaged(ctx, map[string]string{"classbox-data": "/out"}, "stdlib-build", "build", "tests", url)
	if err != nil {
		log.Print(err)
		return &Result{1, []byte("system error during build"), nil}
	}
	return r
}

func BuildBaseline(ctx context.Context) error {
	r, err := Run(ctx, map[string]string{"classbox-data": "/out"}, "stdlib-build", "build", "baseline")
	if err != nil {
		return err
	}
	if !r.Success() {
		return fmt.Errorf("error during baseline build: %v", string(r.Output))
	}
	return nil
}

func BuildMeta(ctx context.Context) (string, error) {
	r, err := RunStaged(ctx, nil, "stdlib-build", "build", "meta")
	if err != nil {
		return "", err
	}
	if len(r.Stages) != 1 {
		return "", fmt.Errorf("error building meta: %v", string(r.Output))
	}
	s := r.Stages[0]
	if !s.Success() {
		return "", fmt.Errorf("error building meta: %v", string(s.Output))
	}

	if !r.Success() {
		return "", fmt.Errorf("error building meta: %v", string(r.Output))
	}
	return s.Output, nil
}

func RunTest(ctx context.Context, name string) *Stage {
	stage := &Stage{Name: fmt.Sprintf("test::%s", name), Test: name}

	r, err := Run(ctx, map[string]string{"classbox-data": "/in"}, "stdlib-run", name+".test", "-test.v")
	if err != nil {
		stage.Status = "exception"
		return stage
	}
	if r.Success() {
		stage.Status = "success"
		return stage
	}
	stage.Status, stage.Output = "failure", string(r.Output)
	return stage
}

func RunPerf(ctx context.Context, name string) (uint64, error) {
	r, _ := Run(ctx, map[string]string{"classbox-data": "/in"},
		"--security-opt", "seccomp=unconfined",
		"stdlib-run",
		"perf", "stat", "-x", ";", "-r", "10",
		name+".test", "-test.run", "Perf")

	var perf uint64
	for _, line := range strings.Split(string(r.Output), "\n") {
		parts := strings.SplitN(line, ";", 3)
		if len(parts) > 2 && strings.HasPrefix(parts[2], "cycles") {
			if s, err := strconv.ParseUint(parts[0], 10, 64); err == nil {
				perf = s
			}
		}
	}
	if perf == 0 {
		return perf, errors.New("perf data not found")
	}
	return perf, nil
}
