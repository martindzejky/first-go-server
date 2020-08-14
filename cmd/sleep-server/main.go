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
	port := "8080"
	route := "/api/sleep"

	log.Println("Seeding the random number generator")
	rand.Seed(time.Now().UnixNano())

	log.Println("Registering the \"" + route + "\" handler")
	http.HandleFunc(route, sleepHandler)

	log.Println("Server listening on port", port)
	log.Fatalln(http.ListenAndServe(":"+port, nil))
}

func sleepHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request")

	// there's a small chance that the request fails
	if rand.Float32() < 0.05 {
		http.Error(res, "Simulated request failure", http.StatusBadRequest)
		return
	}

	query := req.URL.Query()
	min := httpUtils.GetQueryIntValue(query, "min", 100)
	max := httpUtils.GetQueryIntValue(query, "max", 300)

	waitTime := rand.Intn(max-min) + min

	log.Println("Sleeping for", waitTime, "ms")
	time.Sleep(time.Duration(waitTime) * time.Millisecond)

	res.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(res).Encode(responses.SleepResponse{
		Time: waitTime,
	})

	if err != nil {
		log.Fatalln("Error while writing the HTTP response:", err)
	}
}
