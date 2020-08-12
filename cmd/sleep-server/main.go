package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/martindzejky/first-go-server/cmd/common"
	"github.com/martindzejky/first-go-server/internal/http-utils"
)

func main() {
	port := "8080"

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/api/sleep", sleepHandler)

	log.Println("Server listening on port", port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		log.Fatalln("Error while starting the server:", err)
	}
}

func sleepHandler(writer http.ResponseWriter, request *http.Request) {
	log.Println("Received a request")

	query := request.URL.Query()
	min := httpUtils.GetQueryIntValue(query, "min", 100)
	max := httpUtils.GetQueryIntValue(query, "max", 300)

	waitTime := rand.Intn(max-min) + min

	log.Println("Sleeping for", waitTime, "ms")
	time.Sleep(time.Duration(waitTime) * time.Millisecond)

	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(common.SleepResponse{
		Time: waitTime,
	})

	if err != nil {
		log.Fatal("Error while writing the HTTP response", err)
	}
}
