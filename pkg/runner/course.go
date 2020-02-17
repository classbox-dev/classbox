package runner

import (
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/fileutils"
	"github.com/pkg/errors"
	"log"
)

func (rr *Runner) upgradeCourse() error {

	api := rr.apiClient()
	dcl := rr.dockerClient()

	if rr.Docker.Pull {
		log.Print("[INFO] Pulling images...")
		if rr.Docker.Login {
			repo := rr.Docker.Repo
			log.Printf("[INFO] `docker login` for %s", repo.Host)
			if err := dcl.Login(rr.Ctx, repo.Username, repo.Password, repo.Host); err != nil {
				return errors.Wrap(err, "could not login")
			}
		}
		log.Printf("[INFO] `docker pull`")
		if err := dcl.PullImages(rr.Ctx); err != nil {
			return errors.Wrap(err, "could not pull images")
		}
		log.Printf("[INFO] Images updated")
	}

	tests, err := dcl.BuildMeta(rr.Ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := api.UpdateTests(rr.Ctx, tests); err != nil {
		return errors.Wrap(err, "could not save meta")
	}

	if err := dcl.BuildDocs(rr.Ctx, rr.WebURL, rr.DocsURL); err != nil {
		return errors.WithStack(err)
	}

	if err := fileutils.CleanDir(rr.DataDir); err != nil {
		return errors.WithStack(err)
	}

	if err := dcl.BuildBaseline(rr.Ctx); err != nil {
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
		if a.Run == nil {
			return errors.Errorf("run is missing for `%s`, cannot continue", a.Test)
		}
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
