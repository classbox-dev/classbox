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

type ErrorResponse struct {
	Code    int
	Message string `json:"message"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func checkResponse(r *http.Response) error {
	if r.StatusCode/200 == 1 {
		return nil
	}
	e := ErrorResponse{Code: r.StatusCode}
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		return errors.WithStack(err)
	}
	return e
}

func (c *Client) createRequest(ctx context.Context, method string, path string, body []byte) (*http.Request, error) {
	buf := bytes.NewBuffer(body)
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf(c.baseUrl+path), buf)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return req, nil
}

func (c *Client) makeRequest(ctx context.Context, req *http.Request, v interface{}) error {
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

func (c *Client) request(ctx context.Context, method string, path string, body []byte, v interface{}) error {
	req, err := c.createRequest(ctx, method, path, body)
	if err != nil {
		return errors.WithStack(err)
	}
	return c.makeRequest(ctx, req, v)
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

func (c *Client) GetCommit(ctx context.Context, login, commit string) (*models.Commit, error) {
	path := fmt.Sprintf("/commits/%s:%s", login, commit)
	var resp models.Commit
	if err := c.request(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetTests(ctx context.Context) ([]*models.Test, error) {
	var resp []*models.Test
	if err := c.request(ctx, "GET", "/tests", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetOauthUrl(ctx context.Context) (string, error) {
	var resp struct {
		Url string `json:"url"`
	}
	if err := c.request(ctx, "GET", "/signin/oauth", nil, &resp); err != nil {
		return "", err
	}
	return resp.Url, nil
}

func (c *Client) GetUserStats(ctx context.Context) ([]*models.UserStat, error) {
	var resp []*models.UserStat
	if err := c.request(ctx, "GET", "/stats", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
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

func (c *Client) UpdateTests(ctx context.Context, tests []*models.Test) error {
	data, err := json.Marshal(tests)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := c.request(ctx, "PUT", "/tests", data, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetCourse(ctx context.Context) (*models.Course, error) {
	var course models.Course
	if err := c.request(ctx, "GET", "/course", nil, &course); err != nil {
		return nil, err
	}
	return &course, nil
}

func (c *Client) GetUser(ctx context.Context, session string) (*models.User, error) {
	if session == "" {
		return nil, nil
	}
	var user models.User
	req, err := c.createRequest(ctx, "GET", "/user", nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Add("X-Session", session)
	if err := c.makeRequest(ctx, req, &user); err != nil {
		if e, ok := err.(*ErrorResponse); ok && e.Code == http.StatusUnauthorized {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (c *Client) UpdateCourse(ctx context.Context, ready bool) error {
	data, err := json.Marshal(map[string]bool{"is_ready": ready})
	if err != nil {
		return errors.WithStack(err)
	}
	if err := c.request(ctx, "PUT", "/course", data, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateUser(ctx context.Context, code, state string) (*models.AuthStage, error) {
	data, err := json.Marshal(map[string]string{"code": code, "state": state})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var resp models.AuthStage
	if err := c.request(ctx, "POST", "/signin/create", data, &resp); err != nil {
		return nil, errors.WithStack(err)
	}
	return &resp, nil
}

func (c *Client) InstallApp(ctx context.Context, instId uint64, state string) (*models.AuthStage, error) {
	body, err := json.Marshal(&models.AppInstallData{instId, state})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var resp models.AuthStage
	if err := c.request(ctx, "POST", "/signin/install", body, &resp); err != nil {
		return nil, errors.WithStack(err)
	}
	return &resp, nil
}
