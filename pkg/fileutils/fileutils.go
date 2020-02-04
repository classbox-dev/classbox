package fileutils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func Hash(src string) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	//noinspection GoUnhandledErrorResult
	defer in.Close()

	h := sha256.New()
	if _, err := io.Copy(h, in); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func CleanDir(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(path, file.Name())
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

type Artifact struct {
	Path string
	Hash string
}

func SaveArtifacts(dataDir, tmpDir string) (map[string]Artifact, error) {
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("error reading data directory: %w", err)
	}
	tests := map[string]Artifact{}

	for _, file := range files {
		path := filepath.Join(dataDir, file.Name())
		tmpPath := filepath.Join(tmpDir, file.Name())

		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".test") {
			continue
		}

		hash, err := Hash(path)
		if err != nil {
			return nil, fmt.Errorf("hash error: %w", err)
		}

		testName := strings.TrimSuffix(fileName, ".test")
		err = Copy(path, tmpPath)
		if err != nil {
			return nil, fmt.Errorf("copy error: %w", err)
		}

		tests[testName] = Artifact{tmpPath, hash}
	}
	return tests, nil
}
