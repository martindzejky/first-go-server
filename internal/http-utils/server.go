package httpUtils

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/martindzejky/first-go-server/internal/os-signals"
)

func RunServer(port string, handler http.Handler) {
	log.Println("Creating the server on port", port)

	server := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
	}

	signals := osSignals.MakeChannelWithInterruptSignal()

	log.Println("Registering request handlers")

	go func() {
		log.Println("Starting the server")
		err := server.ListenAndServe()

		if err != nil {
			log.Fatalln(err)
		}
	}()

	// block until the interrupt signal is received
	<-signals

	// make a context to wait for connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Gracefully shutting down the server")
	err := server.Shutdown(ctx)

	if err != nil {
		log.Fatalln("Error while shutting down the server:", err)
	}
}
