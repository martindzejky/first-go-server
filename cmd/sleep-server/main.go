package main

import (
	"context"
	"encoding/json"
	osSignals "github.com/martindzejky/first-go-server/internal/os-signals"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/martindzejky/first-go-server/internal/http-utils"
	"github.com/martindzejky/first-go-server/internal/responses"
)

func main() {
	port := "8080"
	router := mux.NewRouter()

	server := httpUtils.MakeServer(port, router)
	signals := osSignals.MakeChannelWithInterruptSignal()

	log.Println("Seeding the random number generator")
	rand.Seed(time.Now().UnixNano())

	log.Println("Registering request handlers")
	router.HandleFunc("/api/sleep", sleepHandler)

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
