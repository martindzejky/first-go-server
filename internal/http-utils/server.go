package httpUtils

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func MakeServer(port string, router *mux.Router) *http.Server {
	log.Println("Creating the server on port", port)

	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	return server
}
