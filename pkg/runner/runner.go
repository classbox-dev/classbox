package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/mkuznets/classbox/pkg/api/client"
	"github.com/mkuznets/classbox/pkg/docker"
	"github.com/mkuznets/classbox/pkg/fileutils"
	"github.com/mkuznets/classbox/pkg/opts"
	"github.com/mkuznets/classbox/pkg/utils"
	"github.com/pkg/errors"
)

type Runner struct {
	Ctx     context.Context
	Http    *http.Client
	Env     *opts.Env
	Jwt     *opts.JwtClient
	Sentry  *opts.Sentry
	Docker  *opts.Docker
	DataDir string
	ApiURL  string
	WebURL  string
	DocsURL string
}

func (rr *Runner) apiClient() *client.Client {
	token, err := rr.Jwt.Token()
	if err != nil {
		panic(err)
	}
	c := client.New(rr.ApiURL)
	c.Auth(token)
	return c
}

func (rr *Runner) dockerClient() *docker.Client {
	return &docker.Client{
		BuilderImage: rr.Docker.BuilderImage,
		RunnerImage:  rr.Docker.RunnerImage,
	}
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
		dockerClient:    rr.dockerClient(),
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
	log.Printf("[INFO] environment: %s", rr.Env.Type)

	upgradeRetries := 0

	for {
		api := rr.apiClient()

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
