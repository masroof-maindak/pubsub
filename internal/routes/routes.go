package routes

import (
	"net/http"

	"github.com/masroof-maindak/pubsub/internal/config"
	"github.com/masroof-maindak/pubsub/internal/handlers"
	"github.com/masroof-maindak/pubsub/internal/utils"
)

func AddRoutes(mux *http.ServeMux, cfg *config.AppConfig) {
	mux.HandleFunc("/connect-pub", handlers.CreateWsConn(utils.Publisher))
	mux.HandleFunc("/connect-sub", handlers.CreateWsConn(utils.Subscriber))
}
