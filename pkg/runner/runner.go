package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mkuznets/classbox/pkg/docker"
	"github.com/mkuznets/classbox/pkg/fileutils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Run struct {
	Hash     string `json:"hash"`
	Status   string `json:"status"`
	Output   string `json:"output,omitempty"`
	Score    uint64 `json:"score"`
	Test     string `json:"test"`
	Baseline bool   `json:"baseline"`
}

type Task struct {
	ID     string `json:"id"`
	Ref    string `json:"ref"`
	URL    string `json:"archive"`
	Stages []docker.Stage
	Runs   []Run
}

type Runner struct {
	Ctx     context.Context
	Http    *http.Client
	ApiURL  string
}

func (rr *Runner) task() (*Task, error) {
	req, err := http.NewRequestWithContext(rr.Ctx, "POST", rr.ApiURL+"/tasks/dequeue", nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	resp, err := rr.Http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send pop request: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, nil
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could read response: %w", err)
	}

	task := Task{}
	err = json.Unmarshal(data, &task)
	if err != nil {
		return nil, fmt.Errorf("could parse response: %w", err)
	}
	return &task, nil
}

func (rr *Runner) finish(task *Task) {
	data, err := json.Marshal(task.Stages)
	if err != nil {
		log.Printf("[ERR] %v", err)
		return
	}

	log.Println(string(data))

	url := fmt.Sprintf("%s/tasks/%s", rr.ApiURL, task.ID)
	req, err := http.NewRequestWithContext(rr.Ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("[ERR] %v", err)
		return
	}
	resp, err := rr.Http.Do(req)
	if err != nil {
		log.Printf("[ERR] [%s] Could not finish task: %v", task.Ref, err)
		return
	}

	log.Println(resp.StatusCode)

	if resp.StatusCode/200 > 1 {
		//noinspection GoUnhandledErrorResult
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		log.Printf("[ERR] [%s] Could not finish task: %v", task.Ref, string(content))
		return
	}

	log.Printf("[INFO] [%s] Finished", task.Ref)
}

func (rr *Runner) event() error {
	task, err := rr.task()
	if err != nil || task == nil {
		return err
	}
	log.Printf("[INFO] [%s] New task id=%s", task.Ref, task.ID)

	defer rr.finish(task)

	dataDir := "/srv/data"
	tmpDir, err := ioutil.TempDir("", "")
	//noinspection GoUnhandledErrorResult
	defer os.RemoveAll(tmpDir)

	err = fileutils.CleanDir(dataDir)
	if err != nil {
		return err
	}

	r := docker.BuildTests(rr.Ctx, task.URL)
	task.Stages = append(task.Stages, r.Stages...)

	log.Printf("[INFO] [%s] Build success: %v\n", task.Ref, r.Success())
	log.Println(r.Stages)

	if !r.Success() {
		return nil
	}

	tests, err := fileutils.SaveArtifacts(dataDir, tmpDir)
	if err != nil {
		return err
	}

	log.Printf("[INFO] [%s] Tests found: %d", task.Ref, len(tests))

	//noinspection GoUnhandledErrorResult
	defer fileutils.CleanDir(dataDir)

	for name, test := range tests {
		err = fileutils.CleanDir(dataDir)
		if err != nil {
			return err
		}

		run := Run{Hash: test.Hash, Test: name, Baseline: false}

		testPath := filepath.Join(dataDir, filepath.Base(test.Path))
		err := fileutils.Copy(test.Path, testPath)
		if err != nil {
			return err
		}
		_ = os.Chmod(testPath, 0500)
		_ = os.Chown(testPath, 2000, 2000)

		s := docker.RunTest(rr.Ctx, name)
		task.Stages = append(task.Stages, *s)
		run.Status = s.Status
		run.Output = s.Output

		log.Printf("[INFO] [%s] Tested `%s`: %s", task.Ref, name, s.Status)

		if s.Success() {
			continue
		}

		// task.GetRuns = append(task.GetRuns, run)
		// _ = docker.RunPerf(rr.Ctx, name)
	}

	return nil
}

func (rr *Runner) upgrade() error {

	meta, err := docker.BuildMeta(rr.Ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(rr.Ctx, "PUT", rr.ApiURL+"/meta", bytes.NewBuffer([]byte(meta)))
	if err != nil {
		return err
	}
	resp, err := rr.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode/200 > 1 {
		//noinspection GoUnhandledErrorResult
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("could not save meta: %v", string(content))
	}

	dataDir := "/srv/data"
	tmpDir, err := ioutil.TempDir("", "")
	//noinspection GoUnhandledErrorResult
	defer os.RemoveAll(tmpDir)

	err = fileutils.CleanDir(dataDir)
	if err != nil {
		return err
	}

	err = docker.BuildBaseline(rr.Ctx)
	if err != nil {
		return err
	}

	tests, err := fileutils.SaveArtifacts(dataDir, tmpDir)
	if err != nil {
		return err
	}

	runs := make([]Run, 0, len(tests))

	//noinspection GoUnhandledErrorResult
	defer fileutils.CleanDir(dataDir)

	for name, test := range tests {
		err = fileutils.CleanDir(dataDir)
		if err != nil {
			return err
		}
		run := Run{Hash: test.Hash, Test: name, Baseline: true}

		testPath := filepath.Join(dataDir, filepath.Base(test.Path))
		err := fileutils.Copy(test.Path, testPath)
		if err != nil {
			return err
		}
		_ = os.Chmod(testPath, 0500)
		_ = os.Chown(testPath, 2000, 2000)

		s := docker.RunTest(rr.Ctx, name)
		if !s.Success() {
			return fmt.Errorf("baseline `%s` fails tests: %v", name, string(s.Output))
		}

		perf, err := docker.RunPerf(rr.Ctx, name)
		if err != nil {
			return fmt.Errorf("could not read perf for %s: %w", name, err)
		}
		run.Score = perf
		run.Status = "success"
		runs = append(runs, run)
	}

	data, err := json.Marshal(runs)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/runs", rr.ApiURL)
	req, err = http.NewRequestWithContext(rr.Ctx, "PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	resp, err = rr.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode/200 > 1 {
		//noinspection GoUnhandledErrorResult
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("could not save runs: %v", string(content))
	}

	return nil
}

func (rr *Runner) Do() {
	// rr.upgrade()
	// return

	for {
		err := rr.event()
		if err != nil {
			log.Printf("[ERR] %v", err)
		}
		time.Sleep(3 * time.Second)
	}
}
