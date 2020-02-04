package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

type Client struct {
	baseUrl string
	http    *http.Client
}

func New(baseUrl string) *Client {
	return &Client{
		baseUrl: baseUrl,
		http:    &http.Client{},
	}
}

type errorResponse struct {
	Code    int
	Message string `json:"message"`
}

func (e *errorResponse) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func checkResponse(r *http.Response) error {
	if r.StatusCode/200 == 1 {
		return nil
	}
	e := errorResponse{Code: r.StatusCode}
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(&e)
}

func (c *Client) request(ctx context.Context, method string, path string, body []byte, v interface{}) error {
	buf := bytes.NewBuffer(body)

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf(c.baseUrl+path), buf)
	if err != nil {
		return errors.WithStack(err)
	}
	r, err := c.http.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	//noinspection GoUnhandledErrorResult
	defer r.Body.Close()

	if err = checkResponse(r); err != nil {
		return err
	}
	if r.StatusCode == http.StatusNoContent || v == nil {
		return nil
	}
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *Client) DequeueTask(ctx context.Context) (*models.Task, error) {
	task := models.Task{}
	if err := c.request(ctx, "POST", "/tasks/dequeue", nil, &task); err != nil {
		return nil, err
	}
	if task.Id == "" {
		return nil, nil
	}
	return &task, nil
}

func (c *Client) FinishTask(ctx context.Context, taskId string, stages []*models.Stage) error {
	path := fmt.Sprintf("/tasks/%s", taskId)
	data, err := json.Marshal(stages)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := c.request(ctx, "POST", path, data, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetRuns(ctx context.Context, hashes []string) (map[string]*models.Run, error) {
	vs := url.Values{}
	for _, h := range hashes {
		vs.Add("hash", h)
	}
	path := fmt.Sprintf("/runs?%s", vs.Encode())

	var runs []*models.Run
	if err := c.request(ctx, "GET", path, nil, &runs); err != nil {
		return nil, err
	}

	m := map[string]*models.Run{}
	for _, r := range runs {
		m[r.Hash] = r
	}
	return m, nil
}

func (c *Client) GetBaselines(ctx context.Context, tests []string) (map[string]*models.Run, error) {
	vs := url.Values{}
	for _, h := range tests {
		vs.Add("test", h)
	}
	path := fmt.Sprintf("/runs/baselines?%s", vs.Encode())

	var runs []*models.Run
	if err := c.request(ctx, "GET", path, nil, &runs); err != nil {
		return nil, err
	}

	m := map[string]*models.Run{}
	for _, r := range runs {
		m[r.Test] = r
	}
	return m, nil
}

func (c *Client) SubmitRuns(ctx context.Context, runs []*models.Run) error {
	data, err := json.Marshal(runs)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := c.request(ctx, "PUT", "/runs", data, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateMeta(ctx context.Context, tests []*models.Test) error {
	data, err := json.Marshal(tests)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := c.request(ctx, "PUT", "/meta", data, nil); err != nil {
		return err
	}
	return nil
}
