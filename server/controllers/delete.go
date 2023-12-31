package controllers

import (
	"github.com/muhtutorials/reminders_cli/server/transport"
	"net/http"
)

type deleter interface {
	Delete(ids []int) error
}

func deleteReminders(service deleter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids, err := parseIDsParam(r.Context())
		if err != nil {
			transport.SendError(w, err)
			return
		}
		err = service.Delete(ids)
		if err != nil {
			transport.SendError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
