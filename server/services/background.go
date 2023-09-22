package services

import (
	"github.com/muhtutorials/reminders_cli/server/models"
	"log"
	"time"
)

type saver interface {
	save() error
}

// BackgroundSaver represents the reminder background saver
type BackgroundSaver struct {
	ticker  *time.Ticker
	done    chan struct{}
	service saver
}

func NewSaver(service saver) *BackgroundSaver {
	ticker := time.NewTicker(30 * time.Second)
	done := make(chan struct{})
	return &BackgroundSaver{
		ticker:  ticker,
		done:    done,
		service: service,
	}
}

func (s BackgroundSaver) Start() {
	log.Println("background saver started")
	for {
		select {
		case <-s.ticker.C:
			err := s.service.save()
			if err != nil {
				log.Printf("could not save records in background: %v", err)
			}
		case <-s.done:
			return
		}
	}
}

func (s BackgroundSaver) Stop() error {
	s.ticker.Stop()
	s.done <- struct{}{}
	err := s.service.save()
	if err != nil {
		return err
	}
	log.Println("background saver stopped")
	return nil
}

// HTTPNotifierClient represents the HTTP client for communicating with the notifier server
type HTTPNotifierClient interface {
	Notify(reminder models.Reminder) (NotificationResponse, error)
}

type snapshotManager interface {
	snapshot() Snapshot
	snapshotGrooming(notifiedReminder ...models.Reminder)
	retry(reminder models.Reminder)
}

// BackgroundNotifier represents the reminder background saver
type BackgroundNotifier struct {
	ticker    *time.Ticker
	done      chan struct{}
	service   snapshotManager
	completed chan models.Reminder
	Client    HTTPNotifierClient
}

func NewNotifier(notifierURL string, service snapshotManager) *BackgroundNotifier {
	ticker := time.NewTicker(time.Second)
	done := make(chan struct{})
	httpClient := NewHTTPClient(notifierURL)
	return &BackgroundNotifier{
		ticker:    ticker,
		done:      done,
		service:   service,
		completed: make(chan models.Reminder),
		Client:    httpClient,
	}
}

func (n BackgroundNotifier) Start() {
	log.Println("background notifier started")
	for {
		select {
		case <-n.ticker.C:
			snapshot := n.service.snapshot()
			for id := range snapshot.Uncompleted {
				_, reminder := snapshot.Uncompleted.flatten(id)
				reminderTick := reminder.ModifiedAt.Add(reminder.Duration).UnixNano()
				nowTick := time.Now().UnixNano()
				deltaTick := time.Now().Add(time.Second).UnixNano()
				if reminderTick > nowTick && reminderTick < deltaTick {
					go n.notify(reminder)
				}
			}
		case r := <-n.completed:
			log.Printf("reminder with: %d was completed\n", r.ID)
		case <-n.done:
			return
		}
	}
}

func (n BackgroundNotifier) Stop() error {
	n.ticker.Stop()
	n.done <- struct{}{}
	log.Println("background notifier stopped")
	return nil
}

// notify notifies a reminder via the HTTP client
func (n BackgroundNotifier) notify(r models.Reminder) {
	res, err := n.Client.Notify(r)
	if err != nil {
		log.Printf("could not notify reminder with id %d\n", r.ID)
		log.Printf("background http client error: %v\n", err)
	} else if res.completed {
		n.service.snapshotGrooming(r)
		n.completed <- r
		return
	}
	n.service.retry(r)
}
