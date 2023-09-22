package main

import (
	"flag"
	"github.com/muhtutorials/reminders_cli/server"
	"github.com/muhtutorials/reminders_cli/server/repositories"
	"github.com/muhtutorials/reminders_cli/server/services"
	"log"
	"os"
	"syscall"
)

func main() {
	var (
		dbFlag          = flag.String("db", "db.json", "Path to db.json file")
		dbCfgFlag       = flag.String("db_cfg", ".db.config.json", "Path to .db.config.json file")
		addrFlag        = flag.String("addr", ":8000", "HTTP server address")
		notifierURLFlag = flag.String("notifier", "http://localhost:5000", "Notifier API URL")
	)
	flag.Parse()

	db := repositories.NewDB(*dbFlag, *dbCfgFlag)
	repo := repositories.NewReminders(db)
	service := services.NewReminders(repo)
	backend := server.NewBackend(*addrFlag, service)
	saver := services.NewSaver(service)
	notifier := services.NewNotifier(*notifierURLFlag, service)

	if err := db.Start(); err != nil {
		log.Fatalf("could not start file database service: %v", err)
	}

	go func() {
		if err := backend.Start(); err != nil {
			log.Fatalf("could not start backend api service: %v", err)
		}
	}()
	go saver.Start()
	go notifier.Start()

	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	server.ListenForSignals(signals, db, backend, saver, notifier)
}
