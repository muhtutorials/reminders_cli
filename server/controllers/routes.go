package controllers

import (
	"github.com/muhtutorials/reminders_cli/server/middleware"
	"net/http"
)

const (
	idParamName  = "id"
	idsParamName = "ids"
	idParam      = "{" + idParamName + "}:^[0-9]+$"
	idsParam     = "{" + idsParamName + "}:[0-9]+(,[0-9]+)*"
)

type RemindersService interface {
	creator
	editor
	fetcher
	deleter
}

type RouterConfig struct {
	Service RemindersService
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := RegexMux{}
	m := middleware.New(middleware.HTTPLogger)
	r.Get("/reminders/"+idsParam, m.Then(fetchReminders(cfg.Service)))
	r.Post("/reminders", m.Then(createReminder(cfg.Service)))
	r.Patch("/reminders/"+idParam, m.Then(editReminder(cfg.Service)))
	r.Delete("/reminders/"+idsParam, m.Then(deleteReminders(cfg.Service)))
	r.Get("/health", m.Then(health()))
	return r
}
