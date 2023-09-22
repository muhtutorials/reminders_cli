package repositories

import (
	"encoding/json"
	"github.com/muhtutorials/reminders_cli/server"
	"github.com/muhtutorials/reminders_cli/server/models"
	"github.com/muhtutorials/reminders_cli/server/services"
	"io"
)

// FileDB represents the file database
type FileDB interface {
	io.ReadWriter
	server.Stopper
	Size() int
	GenerateID() int
}

// Reminders represents the Reminders repository (database layer)
type Reminders struct {
	DB FileDB
}

func NewReminders(db FileDB) *Reminders {
	return &Reminders{
		DB: db,
	}
}

// Save saves the current snapshot of reminders in the DB
func (r Reminders) Save(reminders []models.Reminder) (int, error) {
	bts, err := json.Marshal(reminders)
	if err != nil {
		return 0, err
	}
	n, err := r.DB.Write(bts)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Filter filters reminders by a filtering function
func (r Reminders) Filter(filterFn func(reminder models.Reminder) bool) (services.RemindersMap, error) {
	bts := make([]byte, r.DB.Size())
	n, err := r.DB.Read(bts)
	if err != nil {
		e := models.WrapError("could not read from db", err)
		return services.RemindersMap{}, e
	}
	var reminders []models.Reminder
	err = json.Unmarshal(bts[:n], &reminders)
	if err != nil {
		e := models.WrapError("could not unmarshal json", err)
		return services.RemindersMap{}, e
	}
	rMap := services.RemindersMap{}
	for i, reminder := range reminders {
		if filterFn == nil || filterFn(reminder) {
			reminderMap := map[int]models.Reminder{}
			reminderMap[i] = reminder
			rMap[reminder.ID] = reminderMap
		}
	}
	return rMap, nil
}

// NextID fetches the next DB AUTOINCREMENT id
func (r Reminders) NextID() int {
	return r.DB.GenerateID()
}
