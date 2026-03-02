package routes

import (
	"net/http"

	"github.com/masroof-maindak/pubsub/internal/config"
	"github.com/masroof-maindak/pubsub/internal/handlers"
	"github.com/masroof-maindak/pubsub/internal/utils"
)

func AddRoutes(mux *http.ServeMux, cfg *config.AppConfig) {
	mux.HandleFunc("/ws-conn-pub", handlers.CreateWsConn(utils.Publisher))
	mux.HandleFunc("/ws-conn-sub", handlers.CreateWsConn(utils.Subscriber))
}
