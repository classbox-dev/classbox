package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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

	for {
		time.Sleep(3 * time.Second)
		task, err := w.getTask()
		if err != nil {
			log.Printf("[ERR] %v", err)
		}

		fmt.Println(task)
	}
}
