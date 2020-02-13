package runner

import (
	"context"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/docker"
	"github.com/mkuznets/classbox/pkg/fileutils"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
)

type Artifact struct {
	Test     string
	Path     string
	Hash     string
	Cache    *models.Run
	Baseline *models.Run
	Run      *models.Run
}

type Store struct {
	ref             string
	dataDir         string
	tmpDir          string
	createBaselines bool
	artifacts       []*Artifact
	dockerClient    *docker.Client
}

func (s *Store) Execute(ctx context.Context) error {

	//noinspection GoUnhandledErrorResult
	defer fileutils.CleanDir(s.dataDir) // nolint

	//noinspection GoUnhandledErrorResult
	defer os.RemoveAll(s.tmpDir)

	for _, a := range s.artifacts {

		if a.Cache != nil {
			log.Printf("[INFO] [%s] Using cache for `%v` (hash=%v)", s.ref, a.Test, a.Hash[:16])
			c := *a.Cache
			a.Run = &c
			if !s.createBaselines {
				a.Run.CompareToBaseline(a.Baseline)
			}
			continue
		}

		err := fileutils.CleanDir(s.dataDir)
		if err != nil {
			return errors.WithStack(err)
		}
		testPath := filepath.Join(s.dataDir, filepath.Base(a.Path))
		err = fileutils.Copy(a.Path, testPath)
		if err != nil {
			return errors.WithStack(err)
		}
		_ = os.Chmod(testPath, 0500)
		_ = os.Chown(testPath, 2000, 2000)

		run := &models.Run{Hash: a.Hash}
		err = s.dockerClient.RunTest(ctx, a.Test, run)
		if err != nil {
			log.Printf("[ERR] [%s] error during testing `%s`: %v", s.ref, a.Test, err)
			continue
		}

		log.Printf("[INFO] [%s] `%s` unit tests: %s", s.ref, a.Test, run.Status)

		if run.Status == "success" {
			perf, err := s.dockerClient.RunPerf(ctx, a.Test)
			if err != nil {
				log.Printf("[ERR] [%s] error during perf measuring `%s`: %v", s.ref, a.Test, err)
				continue
			}
			run.Score = perf
			log.Printf("[INFO] [%s] `%s` perf tests: %v", s.ref, a.Test, perf)
		}
		a.Run = run
	}

	if !s.createBaselines {
		for _, a := range s.artifacts {
			if a.Run != nil {
				a.Run.CompareToBaseline(a.Baseline)
			}
		}
	}

	return nil
}
