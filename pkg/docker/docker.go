package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/pkg/errors"
	"os/exec"
	"strconv"
	"strings"
)

type Client struct {
	BuilderImage string
	RunnerImage  string
}

type Result struct {
	ExitCode int
	Output   []byte
	Stages   []*models.Stage
}

func (r *Result) Success() bool {
	return r.ExitCode == 0
}

func (client *Client) runStaged(ctx context.Context, volumes map[string]string, args ...string) (*Result, error) {
	r, err := client.run(ctx, volumes, args...)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(r.Output, &r.Stages)
	if err != nil {
		return nil, fmt.Errorf("run error: %v", string(r.Output))
	}
	return r, nil
}

func (client *Client) run(ctx context.Context, volumes map[string]string, args ...string) (*Result, error) {
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

	var result Result
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			return nil, err
		}
	}
	result.Output = out.Bytes()
	return &result, nil
}

func (client *Client) Login(ctx context.Context, username, password, host string) error {
	cmd := exec.CommandContext(ctx, "docker", "login", "-u", username, "-p", password, host)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, out.String())
	}
	return nil
}

func (client *Client) PullImages(ctx context.Context) error {
	for _, image := range []string{client.BuilderImage, client.RunnerImage} {
		cmd := exec.CommandContext(ctx, "docker", "pull", image)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		if err != nil {
			return errors.Wrap(err, out.String())
		}
	}
	return nil
}

func (client *Client) BuildTests(ctx context.Context, url string) *Result {
	r, err := client.runStaged(ctx, map[string]string{"classbox-data": "/out"}, client.BuilderImage, "build", "tests", url)
	if err != nil {
		return &Result{1, []byte("system error during build"), nil}
	}
	return r
}

func (client *Client) BuildBaseline(ctx context.Context) error {
	r, err := client.run(ctx, map[string]string{"classbox-data": "/out"}, client.BuilderImage, "build", "baseline")
	if err != nil {
		return err
	}
	if !r.Success() {
		return fmt.Errorf("error during baseline build: %v", string(r.Output))
	}
	return nil
}

func (client *Client) BuildDocs(ctx context.Context, webUrl string, docsUrl string) error {
	r, err := client.run(ctx, map[string]string{"classbox-docs": "/out"}, client.BuilderImage, "build", "docs", "--web", webUrl, "--docs", docsUrl)
	if err != nil {
		return err
	}
	if !r.Success() {
		return fmt.Errorf("error during docs build: %v", string(r.Output))
	}
	return nil
}

func (client *Client) BuildMeta(ctx context.Context) ([]*models.Test, error) {
	r, err := client.runStaged(ctx, nil, client.BuilderImage, "build", "meta")
	if err != nil {
		return nil, err
	}
	if !r.Success() {
		return nil, fmt.Errorf("failed to build meta: %v", string(r.Output))
	}
	if len(r.Stages) != 1 {
		return nil, fmt.Errorf("failed to build meta: %v", string(r.Output))
	}
	s := r.Stages[0]
	if !s.Success() {
		return nil, fmt.Errorf("failed to build meta: %v", s.Output)
	}
	var meta []*models.Test
	err = json.Unmarshal([]byte(s.Output), &meta)
	if err != nil {
		return nil, fmt.Errorf("error parting meta: %v", string(r.Output))
	}
	return meta, nil
}

func (client *Client) RunTest(ctx context.Context, test string, run *models.Run) error {
	r, err := client.run(ctx, map[string]string{"classbox-data": "/in"},
		"-e", "TIMEOUT=5", client.RunnerImage,
		test+".test", "-test.v", "-test.run", "Unit",
	)
	if err != nil {
		return err
	}
	if r.Success() {
		run.Status = "success"
	} else {
		run.Status, run.Output = "failure", string(r.Output)
	}
	run.Test = test
	return nil
}

func (client *Client) RunPerf(ctx context.Context, name string) (uint64, error) {
	r, _ := client.run(ctx, map[string]string{"classbox-data": "/in"},
		"--security-opt", "seccomp=unconfined",
		"-e", "TIMEOUT=20",
		client.RunnerImage,
		"perf", "stat", "-x", ";", "-r", "5",
		name+".test", "-test.run", "Perf",
	)

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
