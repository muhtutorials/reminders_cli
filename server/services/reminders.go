package services

import (
	"fmt"
	"github.com/muhtutorials/reminders_cli/server/models"
	"log"
	"strings"
	"time"
)

type RemindersMap map[int]map[int]models.Reminder

// flatten assumes the map has only one value and reads it and retrieves the key and value
func (rm RemindersMap) flatten(id int) (int, models.Reminder) {
	var index int
	var reminder models.Reminder
	for i, r := range rm[id] {
		index = i
		reminder = r
	}
	return index, reminder
}

type ReminderRepository interface {
	Save([]models.Reminder) (int, error)
	Filter(filterFn func(reminder models.Reminder) bool) (RemindersMap, error)
	NextID() int
}

// Snapshot represents current service in memory state
type Snapshot struct {
	All         RemindersMap
	Uncompleted RemindersMap
}

// Reminders represents the Reminders service
type Reminders struct {
	repo     ReminderRepository
	Snapshot Snapshot
}

func NewReminders(repo ReminderRepository) *Reminders {
	return &Reminders{
		repo: repo,
		Snapshot: Snapshot{
			All:         RemindersMap{},
			Uncompleted: RemindersMap{},
		},
	}
}

// Populate populates the reminders service internal state with data from db file
func (rs Reminders) Populate() error {
	all, err := rs.repo.Filter(nil)
	if err != nil {
		return models.WrapError("could not get all reminders", err)
	}
	uncompleted, err := rs.repo.Filter(func(r models.Reminder) bool {
		return r.ModifiedAt.Add(r.Duration).UnixNano() > time.Now().UnixNano()
	})
	if err != nil {
		return models.WrapError("could not get uncompleted reminders", err)
	}
	rs.Snapshot.All = all
	rs.Snapshot.Uncompleted = uncompleted
	return nil
}

// ReminderCreateBody represents the model for creating a reminder
type ReminderCreateBody struct {
	Title       string
	Message     string
	Duration    time.Duration
	RetryPeriod time.Duration
}

func (rs Reminders) Create(body ReminderCreateBody) (models.Reminder, error) {
	nextID := rs.repo.NextID()
	if body.Title == "" {
		err := models.DataValidationError{
			Message: "title cannot be empty",
		}
		return models.Reminder{}, err
	}
	if body.Message == "" {
		err := models.DataValidationError{
			Message: "message cannot be empty",
		}
		return models.Reminder{}, err
	}
	if body.Duration == 0 {
		err := models.DataValidationError{
			Message: "duration cannot be 0",
		}
		return models.Reminder{}, err
	}
	if body.RetryPeriod == 0 {
		err := models.DataValidationError{
			Message: "retry period cannot be 0",
		}
		return models.Reminder{}, err
	}
	reminder := models.Reminder{
		ID:          nextID,
		Title:       body.Title,
		Message:     body.Message,
		Duration:    body.Duration,
		RetryPeriod: body.RetryPeriod,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
	}
	index := len(rs.Snapshot.All)
	rs.Snapshot.All[reminder.ID] = map[int]models.Reminder{index: reminder}
	rs.Snapshot.Uncompleted[reminder.ID] = map[int]models.Reminder{index: reminder}
	return reminder, nil
}

type ReminderEditBody struct {
	ID          int
	Title       string
	Message     string
	Duration    time.Duration
	RetryPeriod time.Duration
}

func (rs Reminders) Edit(reminderBody ReminderEditBody) (models.Reminder, error) {
	_, ok := rs.Snapshot.All[reminderBody.ID]
	if !ok {
		err := models.NotFoundError{
			Message: fmt.Sprintf("could not find reminder with id: %d", reminderBody.ID),
		}
		return models.Reminder{}, err
	}
	changed := false
	index, reminder := rs.Snapshot.All.flatten(reminderBody.ID)
	if strings.TrimSpace(reminderBody.Title) != "" {
		reminder.Title = reminderBody.Title
		changed = true
	}
	if strings.TrimSpace(reminderBody.Message) != "" {
		reminder.Message = reminderBody.Message
		changed = true
	}
	if reminderBody.Duration != 0 {
		reminder.Duration = reminderBody.Duration
		changed = true
	}
	if reminderBody.RetryPeriod != 0 {
		reminder.RetryPeriod = reminderBody.RetryPeriod
		changed = true
	}
	if !changed {
		err := models.FormatValidationError{
			Message: "body must contain at least 1 of: 'title', 'message', 'duration', 'retryPeriod'",
		}
		return models.Reminder{}, err
	}
	reminder.ModifiedAt = time.Now()
	rs.Snapshot.All[reminder.ID] = map[int]models.Reminder{index: reminder}
	if reminder.ModifiedAt.UnixNano() < time.Now().Add(reminderBody.Duration).UnixNano() {
		rs.Snapshot.Uncompleted[reminder.ID] = map[int]models.Reminder{index: reminder}
	} else {
		delete(rs.Snapshot.Uncompleted, reminder.ID)
	}
	return reminder, nil
}

func (rs Reminders) Fetch(ids []int) ([]models.Reminder, error) {
	reminders := make([]models.Reminder, 0)
	var notFound []int
	for _, id := range ids {
		_, ok := rs.Snapshot.All[id]
		if !ok {
			notFound = append(notFound, id)
			continue
		}
		_, reminder := rs.Snapshot.All.flatten(id)
		reminders = append(reminders, reminder)
	}
	if len(notFound) > 0 {
		err := models.NotFoundError{
			Message: fmt.Sprintf("could not find reminders with ids: %v", notFound),
		}
		return []models.Reminder{}, err
	}
	return reminders, nil
}

func (rs Reminders) Delete(ids []int) error {
	var notFound []int
	for _, id := range ids {
		_, ok := rs.Snapshot.All[id]
		if !ok {
			notFound = append(notFound, id)
		}
	}
	if len(notFound) > 0 {
		return models.NotFoundError{
			Message: fmt.Sprintf("could not find reminders with ids: %v", notFound),
		}
	}
	for _, id := range ids {
		delete(rs.Snapshot.All, id)
		delete(rs.Snapshot.Uncompleted, id)
	}
	return nil
}

func (rs Reminders) save() error {
	reminders := make([]models.Reminder, len(rs.Snapshot.All))
	for _, remindersMap := range rs.Snapshot.All {
		for i, reminder := range remindersMap {
			reminders[i] = reminder
		}
	}
	n, err := rs.repo.Save(reminders)
	if err != nil {
		return models.WrapError("could not save snapshot", err)
	}
	if n > 0 && len(reminders) != 0 {
		log.Printf("successfully saved snapshot: %d reminders", len(reminders))
	}
	return nil
}

// GetSnapshot fetches the current service snapshot
func (rs Reminders) snapshot() Snapshot {
	return rs.Snapshot
}

// snapshotGrooming clears the current snapshot from notified reminders
func (rs Reminders) snapshotGrooming(notifiedReminders ...models.Reminder) {
	if len(notifiedReminders) > 0 {
		log.Printf("snapshot grooming: %d record(s)", len(notifiedReminders))
	}
	for _, reminder := range notifiedReminders {
		delete(rs.Snapshot.Uncompleted, reminder.ID)
		reminder.Duration = -time.Hour
		index, _ := rs.Snapshot.All.flatten(reminder.ID)
		rs.Snapshot.All[reminder.ID] = map[int]models.Reminder{index: reminder}
	}
}

// retry retries a reminder by resetting its duration
func (rs Reminders) retry(reminder models.Reminder) {
	reminder.ModifiedAt = time.Now()
	reminder.Duration = reminder.RetryPeriod

	log.Printf(
		"retrying record with id: %d after %v",
		reminder.ID,
		reminder.Duration.String(),
	)
	index, _ := rs.Snapshot.All.flatten(reminder.ID)
	rs.Snapshot.All[reminder.ID] = map[int]models.Reminder{index: reminder}
	rs.Snapshot.Uncompleted[reminder.ID] = map[int]models.Reminder{index: reminder}
}
