package runner

import (
	"context"
	"fmt"
	"github.com/mkuznets/classbox/pkg/api/client"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/docker"
	"github.com/mkuznets/classbox/pkg/fileutils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Runner struct {
	Ctx    context.Context
	Http   *http.Client
	ApiURL string
}

func (rr *Runner) apiClient() *client.Client {
	return client.New(rr.ApiURL)
}

func (rr *Runner) finish(task *models.Task) {
	api := rr.apiClient()
	err := api.SubmitRuns(rr.Ctx, task.Runs)
	if err != nil {
		log.Printf("[WARN] [%s] Could not submit runs: %v", task.Ref, err)
	}
	err = api.FinishTask(rr.Ctx, task.Id, task.Stages)
	if err != nil {
		log.Printf("[ERR] [%s] Could not finish task: %v", task.Ref, err)
		return
	}
	log.Printf("[INFO] [%s] Finished", task.Ref)
}

func (rr *Runner) event() error {

	api := rr.apiClient()
	task, err := api.DequeueTask(rr.Ctx)

	if err != nil || task == nil {
		return err
	}
	log.Printf("[INFO] [%s] New task id=%s", task.Ref, task.Id)

	defer rr.finish(task)

	dataDir := "/srv/data"
	tmpDir, err := ioutil.TempDir("", "")
	//noinspection GoUnhandledErrorResult
	defer os.RemoveAll(tmpDir)

	err = fileutils.CleanDir(dataDir)
	if err != nil {
		return err
	}

	r := docker.BuildTests(rr.Ctx, task.Url)
	task.Stages = append(task.Stages, r.Stages...)

	log.Printf("[INFO] [%s] Build success: %v\n", task.Ref, r.Success())

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

	var hashes, testsNames []string
	for name, test := range tests {
		hashes = append(hashes, test.Hash)
		testsNames = append(testsNames, name)
	}

	cachedRuns, err := api.GetRuns(rr.Ctx, hashes)
	if err != nil {
		log.Printf("[ERR] [%s] Could not get cached runs: %v", task.Ref, err)
	}

	baselines, err := api.GetBaselines(rr.Ctx, testsNames)
	if err != nil {
		task.ReportSystemError("")
		return err
	}

	for name, test := range tests {

		if cr, ok := cachedRuns[test.Hash]; ok {
			log.Printf("[INFO] [%s] Using cache for %v (hash=%v)", task.Ref, name, test.Hash)
			stage := models.Stage{}
			stage.FillFromRun("test", cr)
			task.Stages = append(task.Stages, &stage)
			continue
		}

		baseline, ok := baselines[name]
		if !ok {
			log.Printf("[ERR] [%s] baseline for `%s` is not found", task.Ref, name)
			task.ReportSystemError(name)
			continue
		}

		err = fileutils.CleanDir(dataDir)
		if err != nil {
			return err
		}
		testPath := filepath.Join(dataDir, filepath.Base(test.Path))
		err := fileutils.Copy(test.Path, testPath)
		if err != nil {
			return err
		}
		_ = os.Chmod(testPath, 0500)
		_ = os.Chown(testPath, 2000, 2000)

		run := &models.Run{Hash: test.Hash, Baseline: false}
		err = docker.RunTest(rr.Ctx, name, run)
		if err != nil {
			log.Printf("[ERR] [%s] error during testing `%s`: %v", task.Ref, name, err)
			task.ReportSystemError(name)
			continue
		}

		log.Printf("[INFO] [%s] Tested `%s`: %s", task.Ref, name, run.Status)

		if run.Status == "success" {
			perf, err := docker.RunPerf(rr.Ctx, name)
			if err != nil {
				log.Printf("[ERR] [%s] error during perf measuring `%s`: %v", task.Ref, name, err)
				task.ReportSystemError(name)
				continue
			}
			run.Score = perf
			percent := run.Score * 100 / (baseline.Score * 5 / 4)
			run.Output = fmt.Sprintf("you are at %v%% of the baseline", percent)

			if percent > 100 {
				run.Status = "failure"
			}
		}

		task.Runs = append(task.Runs, run)
		stage := &models.Stage{}
		stage.FillFromRun("test", run)
		task.Stages = append(task.Stages, stage)
	}

	return nil
}

// func (rr *Runner) upgrade() error {
//
// 	meta, err := docker.BuildMeta(rr.Ctx)
// 	if err != nil {
// 		return err
// 	}
//
// 	req, err := http.NewRequestWithContext(rr.Ctx, "PUT", rr.ApiURL+"/meta", bytes.NewBuffer([]byte(meta)))
// 	if err != nil {
// 		return err
// 	}
// 	resp, err := rr.Http.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	if resp.StatusCode/200 > 1 {
// 		//noinspection GoUnhandledErrorResult
// 		defer resp.Body.Close()
// 		content, _ := ioutil.ReadAll(resp.Body)
// 		return fmt.Errorf("could not save meta: %v", string(content))
// 	}
//
// 	dataDir := "/srv/data"
// 	tmpDir, err := ioutil.TempDir("", "")
// 	//noinspection GoUnhandledErrorResult
// 	defer os.RemoveAll(tmpDir)
//
// 	err = fileutils.CleanDir(dataDir)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = docker.BuildBaseline(rr.Ctx)
// 	if err != nil {
// 		return err
// 	}
//
// 	tests, err := fileutils.SaveArtifacts(dataDir, tmpDir)
// 	if err != nil {
// 		return err
// 	}
//
// 	runs := make([]*models.Run, 0, len(tests))
//
// 	//noinspection GoUnhandledErrorResult
// 	defer fileutils.CleanDir(dataDir)
//
// 	for name, test := range tests {
// 		err = fileutils.CleanDir(dataDir)
// 		if err != nil {
// 			return err
// 		}
// 		run := models.Run{Hash: test.Hash, Test: name, Baseline: true}
//
// 		testPath := filepath.Join(dataDir, filepath.Base(test.Path))
// 		err := fileutils.Copy(test.Path, testPath)
// 		if err != nil {
// 			return err
// 		}
// 		_ = os.Chmod(testPath, 0500)
// 		_ = os.Chown(testPath, 2000, 2000)
//
// 		s := docker.RunTest(rr.Ctx, name)
// 		if !s.Success() {
// 			return fmt.Errorf("baseline `%s` fails tests: %v", name, s.Output)
// 		}
//
// 		perf, err := docker.RunPerf(rr.Ctx, name)
// 		if err != nil {
// 			return fmt.Errorf("could not read perf for %s: %w", name, err)
// 		}
// 		run.Score = perf
// 		run.Status = "success"
// 		runs = append(runs, &run)
// 	}
//
// 	data, err := json.Marshal(runs)
// 	if err != nil {
// 		return err
// 	}
//
// 	url := fmt.Sprintf("%s/runs", rr.ApiURL)
// 	req, err = http.NewRequestWithContext(rr.Ctx, "PUT", url, bytes.NewBuffer(data))
// 	if err != nil {
// 		return err
// 	}
// 	resp, err = rr.Http.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	if resp.StatusCode/200 > 1 {
// 		//noinspection GoUnhandledErrorResult
// 		defer resp.Body.Close()
// 		content, _ := ioutil.ReadAll(resp.Body)
// 		return fmt.Errorf("could not save runs: %v", string(content))
// 	}
//
// 	return nil
// }

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
