package models

import "time"

type Reminder struct {
	ID          int           `json:"id"`
	Title       string        `json:"title"`
	Message     string        `json:"message"`
	Duration    time.Duration `json:"duration"`
	RetryPeriod time.Duration `json:"retry_period"`
	CreatedAt   time.Time     `json:"created_at"`
	ModifiedAt  time.Time     `json:"modified_at"`
}
