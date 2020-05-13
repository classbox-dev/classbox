package runner

import (
	"fmt"
	"log"

	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/fileutils"
	"github.com/pkg/errors"
)

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
	dcl := rr.dockerClient()

	err := fileutils.CleanDir(rr.DataDir)
	if err != nil {
		return err
	}

	r := dcl.BuildTests(rr.Ctx, task.Url)
	task.Stages = append(task.Stages, r.Stages...)

	log.Printf("[INFO] [%s] build completed", task.Ref)

	store, err := rr.newStore(task.Ref, false)
	if err != nil {
		task.ReportSystemError("")
		return errors.WithStack(err)
	}

	// determine cached builds
	builds := make(map[string]*models.Stage)
	for _, s := range task.Stages {
		var test string
		n, err := fmt.Sscanf(s.Name, "build::%s", &test)
		if err == nil && n == 1 {
			builds[test] = s
		}
	}
	for _, a := range store.artifacts {
		st, ok := builds[a.Test]
		if !ok {
			continue
		}
		if a.Cache != nil {
			st.Cached = true
		}
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
		stage := &models.Stage{
			Cached: a.Cache != nil,
			Run:    &models.RunHash{Hash: a.Run.Hash},
		}
		stage.FillFromRun("test", a.Run)
		task.Stages = append(task.Stages, stage)
	}
	return nil
}
