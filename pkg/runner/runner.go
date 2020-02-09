package runner

import (
	"context"
	"fmt"
	"github.com/mkuznets/classbox/pkg/api/client"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/docker"
	"github.com/mkuznets/classbox/pkg/fileutils"
	"github.com/mkuznets/classbox/pkg/utils"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Runner struct {
	Ctx     context.Context
	Http    *http.Client
	DataDir string
	ApiURL  string
	WebURL  string
	DocsURL string
}

func (rr *Runner) apiClient() *client.Client {
	return client.New(rr.ApiURL)
}

func (rr *Runner) finishTask(task *models.Task) {
	api := rr.apiClient()
	if err := api.SubmitRuns(rr.Ctx, task.Runs); err != nil {
		log.Printf("[WARN] [%s] could not submit runs: %v", task.Ref, err)
	}
	if err := api.FinishTask(rr.Ctx, task.Id, task.Stages); err != nil {
		log.Printf("[ERR] [%s] could not finish task: %v", task.Ref, err)
		return
	}
	log.Printf("[INFO] [%s] finished", task.Ref)
}

func (rr *Runner) runTask(task *models.Task) error {

	err := fileutils.CleanDir(rr.DataDir)
	if err != nil {
		return err
	}

	r := docker.BuildTests(rr.Ctx, task.Url)
	task.Stages = append(task.Stages, r.Stages...)

	log.Printf("[INFO] [%s] build completed", task.Ref)

	store, err := rr.newStore(task.Ref, false)
	if err != nil {
		task.ReportSystemError("")
		return errors.WithStack(err)
	}

	log.Printf("[INFO] [%s] tests found: %d", task.Ref, len(store.artifacts))
	err = store.Execute(rr.Ctx)
	if err != nil {
		task.ReportSystemError("")
		return errors.WithStack(err)
	}

	for _, a := range store.artifacts {
		if a.Run == nil {
			task.ReportSystemError(a.Test)
			continue
		}
		task.Runs = append(task.Runs, a.Run)
		stage := &models.Stage{}
		stage.FillFromRun("test", a.Run)
		task.Stages = append(task.Stages, stage)
	}
	return nil
}

func (rr *Runner) upgradeCourse() error {

	api := rr.apiClient()

	tests, err := docker.BuildMeta(rr.Ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := api.UpdateTests(rr.Ctx, tests); err != nil {
		return errors.Wrap(err, "could not save meta")
	}

	if err := docker.BuildDocs(rr.Ctx, rr.WebURL, rr.DocsURL); err != nil {
		return errors.WithStack(err)
	}

	if err := fileutils.CleanDir(rr.DataDir); err != nil {
		return errors.WithStack(err)
	}

	if err := docker.BuildBaseline(rr.Ctx); err != nil {
		return errors.WithStack(err)
	}

	store, err := rr.newStore("upgrade", true)
	if err != nil {
		return errors.WithStack(err)
	}

	err = store.Execute(rr.Ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	runs := make([]*models.Run, 0, len(store.artifacts))

	for _, a := range store.artifacts {
		r := *a.Run
		r.Baseline = true
		runs = append(runs, &r)
	}

	if err := api.SubmitRuns(rr.Ctx, runs); err != nil {
		return errors.WithStack(err)
	}

	if err := api.UpdateCourse(rr.Ctx, true); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (rr *Runner) newStore(ref string, createBaselines bool) (*Store, error) {

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	st := &Store{
		ref:             ref,
		dataDir:         rr.DataDir,
		tmpDir:          tmpDir,
		createBaselines: createBaselines,
	}

	files, err := ioutil.ReadDir(rr.DataDir)
	if err != nil {
		return nil, fmt.Errorf("error reading source directory: %w", err)
	}

	for _, file := range files {
		path := filepath.Join(rr.DataDir, file.Name())
		tmpPath := filepath.Join(tmpDir, file.Name())

		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".test") {
			continue
		}

		hash, err := fileutils.Hash(path)
		if err != nil {
			return nil, fmt.Errorf("hash error: %w", err)
		}

		testName := strings.TrimSuffix(fileName, ".test")
		err = fileutils.Copy(path, tmpPath)
		if err != nil {
			return nil, fmt.Errorf("copy error: %w", err)
		}

		st.artifacts = append(st.artifacts, &Artifact{
			Test:     testName,
			Path:     tmpPath,
			Hash:     hash,
			Cache:    nil,
			Baseline: nil,
		})
	}

	api := rr.apiClient()

	cachedRuns, err := api.GetRuns(rr.Ctx, utils.UniqueStringFields(st.artifacts, "Hash"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, a := range st.artifacts {
		if v, ok := cachedRuns[a.Hash]; ok {
			a.Cache = v
		}
	}

	if !createBaselines {
		baselines, err := api.GetBaselines(rr.Ctx, utils.UniqueStringFields(st.artifacts, "Test"))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		for _, a := range st.artifacts {
			a.Baseline = baselines[a.Test]
		}
	}

	return st, nil
}

func (rr *Runner) Do() {
	api := rr.apiClient()

	upgradeRetries := 0

	for {
		err := func() error {
			if upgradeRetries > 2 {
				return nil
			}
			course, err := api.GetCourse(rr.Ctx)
			if err != nil {
				return err
			}
			if course.Ready {
				return nil
			}
			log.Printf("[INFO] course upgrade: started")
			err = rr.upgradeCourse()
			if err != nil {
				return err
			}
			log.Printf("[INFO] course upgrade: done")
			upgradeRetries = 0
			return nil
		}()

		if err != nil {
			log.Printf("[WARN] could not upgrade course: %v", err)
			upgradeRetries++
		}

		func() {
			task, err := api.DequeueTask(rr.Ctx)
			if err != nil {
				log.Printf("[ERR] could not dequeue task: %v", err)
				return
			}
			if task == nil {
				return
			}

			log.Printf("[INFO] [%s] new task id=%s", task.Ref, task.Id)
			defer rr.finishTask(task)

			err = rr.runTask(task)
			if err != nil {
				log.Printf("[ERR] [%s] execution error: %v", task.Ref, err)
			}
		}()

		time.Sleep(3 * time.Second)
	}
}
