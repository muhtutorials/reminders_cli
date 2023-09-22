package services

import (
	"bytes"
	"encoding/json"
	"github.com/muhtutorials/reminders_cli/server/models"
	"io"
	"net/http"
	"time"
)

type HTTPClient struct {
	notifierURL string
	client      *http.Client
}

func NewHTTPClient(url string) HTTPClient {
	return HTTPClient{
		notifierURL: url,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// NotificationResponse represents OS notification response for background notifier
type NotificationResponse struct {
	completed   bool
	retryPeriod time.Duration
}

// Notify pushes a given reminder to the notifier service
// if the reminder is nil, means the record must be retried
func (h HTTPClient) Notify(reminder models.Reminder) (NotificationResponse, error) {
	var notifierResponse struct {
		Action string `json:"action"`
	}

	bts, err := json.Marshal(reminder)
	if err != nil {
		e := models.WrapError("could not marshal json", err)
		return NotificationResponse{}, e
	}

	res, err := h.client.Post(h.notifierURL+"/notify", "application/json", bytes.NewReader(bts))
	if err != nil {
		e := models.WrapError("notifier service is unavailable", err)
		return NotificationResponse{}, e
	}

	err = json.NewDecoder(res.Body).Decode(&notifierResponse)
	if err != nil && err != io.EOF {
		e := models.WrapError("could not decode notifier response", err)
		return NotificationResponse{}, e
	}

	action := notifierResponse.Action
	if action == "dismissed" {
		return NotificationResponse{completed: true}, nil
	}

	return NotificationResponse{}, nil
}
