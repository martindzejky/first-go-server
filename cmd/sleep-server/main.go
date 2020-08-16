package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/martindzejky/first-go-server/internal/http-utils"
	"github.com/martindzejky/first-go-server/internal/responses"
)

func main() {
	port := "8080"
	router := mux.NewRouter()

	log.Println("Creating the server")
	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	log.Println("Seeding the random number generator")
	rand.Seed(time.Now().UnixNano())

	log.Println("Registering request handlers")
	router.HandleFunc("/api/sleep", sleepHandler)
	http.Handle("/", router)

	go func() {
		log.Println("Starting the server on port", port)
		err := server.ListenAndServe()

		if err != nil {
			log.Fatalln(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// block until the signal is received
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

func sleepHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request")

	// there's a small chance that the request fails
	if rand.Float32() < 0.05 {
		http.Error(res, "Simulated request failure", http.StatusInternalServerError)
		return
	}

	// get the sleeping time
	query := req.URL.Query()
	min := httpUtils.GetQueryIntValue(query, "min", 100)
	max := httpUtils.GetQueryIntValue(query, "max", 300)

	sleepTime := rand.Intn(max-min) + min

	log.Println("Sleeping for", sleepTime, "ms")

	select {
	case <-req.Context().Done():
		log.Println("Request canceled:", req.Context().Err())
		return
	case <-time.After(time.Duration(sleepTime) * time.Millisecond):
	}

	res.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(res).Encode(responses.SleepResponse{
		Time: sleepTime,
	})

	if err != nil {
		log.Fatalln("Error while writing the HTTP response:", err)
	}
}
