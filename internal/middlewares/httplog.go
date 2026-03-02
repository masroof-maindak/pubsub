package middlewares

import (
	"net/http"

	"github.com/MadAppGang/httplog"
	"github.com/gorilla/websocket"
)

func ConditionalLogger(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// If websocket connection, skip `httplogger`
		// TODO: use variable for route
		if r.URL.Path == "/connect" && websocket.IsWebSocketUpgrade(r) {
			next.ServeHTTP(w, r)
			return
		}

		h := httplog.Logger(next)
		h.ServeHTTP(w, r)
	}
}
