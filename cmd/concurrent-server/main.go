package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/martindzejky/first-go-server/cmd/common"
	"github.com/martindzejky/first-go-server/internal/http-utils"
)

func main() {
	port := "8081"

	// register listeners
	http.HandleFunc("/api/all", apiAllRoute)

	// start the server
	log.Print("Starting the server on port ", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handles the /api/all route - it waits for and returns all 3 responses
func apiAllRoute(res http.ResponseWriter, req *http.Request) {
	log.Print("Received a request for /api/all")

	httpUtils.GetQueryIntValue(req.URL.Query(), "timeout", 1000)

	sleepRes, err := makeRequestToSleepServer()

	if err != nil {
		res.WriteHeader(500)
		log.Fatalln("Error from the sleep server:", err)
	}

	res.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(res).Encode(common.AllResponse{
		Times: []int{sleepRes},
	})
}

// makes a request to the sleeping server
func makeRequestToSleepServer() (int, error) {
	res, err := http.Get("http://localhost:8080/api/sleep")

	if err != nil {
		return 0, err
	}

	if res.StatusCode != 200 {
		return 0, errors.New("The sleep server returned an error: " + res.Status)
	}

	var data common.SleepResponse
	err = json.NewDecoder(res.Body).Decode(&data)

	if err != nil {
		return 0, err
	}

	return data.Time, nil
}
