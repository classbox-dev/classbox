package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func New(ctx context.Context, url string) *Worker {
	w := &Worker{ctx, &http.Client{}, url}
	go w.do()
	return w
}

type Worker struct {
	ctx    context.Context
	http   *http.Client
	apiURL string
}

type taskData struct {
	Id      string `json:"id"`
	Login   string `json:"login"`
	Commit  string `json:"commit"`
	Archive string `json:"archive"`
}

func (w *Worker) getTask() (*taskData, error) {
	req, err := http.NewRequestWithContext(w.ctx, "POST", w.apiURL+"/queue/pop", nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}
	resp, err := w.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send pop request: %w", err)
	}

	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could read response: %w", err)
	}

	task := taskData{}
	err = json.Unmarshal(data, &task)
	if err != nil {
		return nil, fmt.Errorf("could parse response: %w", err)
	}
	return &task, nil
}

func (w *Worker) do() {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}

	for {
		time.Sleep(3 * time.Second)
		task, err := w.getTask()
		if err != nil {
			log.Printf("[ERR] %v", err)
			continue
		}

		fmt.Println(task)
	}
}
