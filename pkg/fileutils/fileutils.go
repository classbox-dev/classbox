package fileutils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
