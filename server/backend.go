package server

import (
	"context"
	"fmt"
	"github.com/muhtutorials/reminders_cli/server/controllers"
	"github.com/muhtutorials/reminders_cli/server/models"
	"github.com/muhtutorials/reminders_cli/server/services"
	"log"
	"net/http"
	"time"
)

type Backend struct {
	server  *http.Server
	service *services.Reminders
}

func NewBackend(addr string, service *services.Reminders) *Backend {
	cfg := controllers.RouterConfig{Service: service}
	router := controllers.NewRouter(cfg)
	return &Backend{
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		service: service,
	}
}

// Start starts the initialized server (backend) application
func (b *Backend) Start() error {
	log.Println("application started on address", b.server.Addr)
	err := b.service.Populate()
	if err != nil {
		return models.WrapError("could not initialize reminders service", err)
	}

	err = b.server.ListenAndServe()
	if err == http.ErrServerClosed {
		log.Println("http server is closed")
		return nil
	}
	return err
}

// Stop gracefully stops the server (backend) application
func (b *Backend) Stop() error {
	timeout := 2 * time.Second
	done, err := make(chan struct{}), make(chan error)

	go func() {
		log.Println("shutting down the http server")
		if e := b.server.Shutdown(context.Background()); e != nil {
			err <- models.WrapError("error on server shutdown", e)
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
		log.Println("application was shut down")
		return nil
	case e := <-err:
		return e
	case <-time.After(timeout):
		return fmt.Errorf("shudown timeout of %v", timeout)
	}
}
