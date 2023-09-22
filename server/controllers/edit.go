package controllers

import (
	"encoding/json"
	"github.com/muhtutorials/reminders_cli/server/models"
	"github.com/muhtutorials/reminders_cli/server/services"
	"github.com/muhtutorials/reminders_cli/server/transport"
	"net/http"
	"time"
)

type editor interface {
	Edit(reminderBody services.ReminderEditBody) (models.Reminder, error)
}

func editReminder(service editor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r.Context())
		if err != nil {
			transport.SendError(w, err)
		}
		var body struct {
			Title       string        `json:"title"`
			Message     string        `json:"message"`
			Duration    time.Duration `json:"duration"`
			RetryPeriod time.Duration `json:"retry_period"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			transport.SendError(w, models.InvalidJSONError{Message: err.Error()})
		}
		reminder, err := service.Edit(services.ReminderEditBody{
			ID:          id,
			Title:       body.Title,
			Message:     body.Message,
			Duration:    body.Duration,
			RetryPeriod: body.RetryPeriod,
		})
		if err != nil {
			transport.SendError(w, err)
			return
		}
		transport.SendJSON(w, reminder, http.StatusOK)
	})
}
