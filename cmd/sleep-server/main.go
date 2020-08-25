package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/martindzejky/first-go-server/internal/http-utils"
	"github.com/martindzejky/first-go-server/internal/responses"
)

func main() {
	log.Println("Seeding the random number generator")
	rand.Seed(time.Now().UnixNano())

	port := "8080"
	mux := http.NewServeMux()

	mux.HandleFunc("/api/sleep", sleepHandler)

	httpUtils.RunServer(port, mux)
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
	max := httpUtils.GetQueryIntValue(query, "max", 600)

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
