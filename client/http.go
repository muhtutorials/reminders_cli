package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient struct {
	client     *http.Client
	BackendURL string
}

type reminderBody struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Message     string        `json:"message"`
	Duration    time.Duration `json:"duration"`
	RetryPeriod time.Duration `json:"retry_period"`
}

func NewHTTPClient(url string) HTTPClient {
	return HTTPClient{
		BackendURL: url,
		client:     &http.Client{},
	}
}

func (c HTTPClient) Create(title, message string, duration, retryPeriod time.Duration) ([]byte, error) {
	requestBody := reminderBody{
		Title:       title,
		Message:     message,
		Duration:    duration,
		RetryPeriod: retryPeriod,
	}

	return c.apiCall(http.MethodPost, "/reminders", &requestBody, http.StatusCreated)
}

func (c HTTPClient) Edit(id, title, message string, duration, retryPeriod time.Duration) ([]byte, error) {
	requestBody := reminderBody{
		ID:          id,
		Title:       title,
		Message:     message,
		Duration:    duration,
		RetryPeriod: retryPeriod,
	}
	return c.apiCall(http.MethodPatch, "/reminders/"+id, &requestBody, http.StatusOK)
}

func (c HTTPClient) Fetch(ids []string) ([]byte, error) {
	idsStr := strings.Join(ids, ",")
	return c.apiCall(http.MethodGet, "/reminders/"+idsStr, nil, http.StatusOK)
}

func (c HTTPClient) Delete(ids []string) error {
	idsStr := strings.Join(ids, ",")
	_, err := c.apiCall(http.MethodDelete, "/reminders/"+idsStr, nil, http.StatusNoContent)
	return err
}

func (c HTTPClient) Healthy(host string) bool {
	res, err := http.Get(host + "/health")
	if err != nil || res.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func (c HTTPClient) apiCall(method, path string, body any, resCode int) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		e := wrapError("could not marshal request body", err)
		return nil, e
	}

	req, err := http.NewRequest(method, c.BackendURL+path, bytes.NewReader(data))
	if err != nil {
		e := wrapError("could not create new request", err)
		return nil, e
	}

	res, err := c.client.Do(req)
	if err != nil {
		e := wrapError("could not make http call", err)
		return nil, e
	}

	resBody, err := c.readResBody(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != resCode {
		if len(resBody) > 0 {
			fmt.Printf("got this response body:\n%s\n", resBody)
		}
		return nil, fmt.Errorf(
			"expected response code: %d, got %d",
			resCode,
			res.StatusCode,
		)
	}

	return []byte(resBody), nil
}

func (c HTTPClient) readResBody(b io.Reader) (string, error) {
	data, err := io.ReadAll(b)
	if err != nil {
		return "", wrapError("could not read response body", err)
	}

	if len(data) == 0 {
		return "", nil
	}

	var buf bytes.Buffer

	if err := json.Indent(&buf, data, "", "\t"); err != nil {
		return "", wrapError("could not indent json", err)
	}

	return buf.String(), nil
}
