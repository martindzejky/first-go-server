package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/martindzejky/first-go-server/cmd/common"
	"github.com/martindzejky/first-go-server/internal/http-utils"
)

func main() {
	port := "8081"

	// register listeners
	log.Println("Registering listeners")
	http.HandleFunc("/api/all", apiAllRoute)

	// start the server
	log.Println("Starting the server on port", port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		log.Fatalln("Failed to listen on the port:", err)
	}
}

// handles the /api/all route - it waits for and returns all 3 responses
func apiAllRoute(res http.ResponseWriter, req *http.Request) {
	log.Print("Received a request for 'all'")

	timeout := httpUtils.GetQueryIntValue(req.URL.Query(), "timeout", 1000)
	timeoutChannel := make(chan bool, 1)
	dataChannel := make(chan int, 3)
	errorsChannel := make(chan error, 3)

	values := make([]int, 0)

	// make the 3 requests
	for i := 0; i < 3; i++ {
		log.Println("Making request n.", i)
		go makeRequestToSleepServer(dataChannel, errorsChannel)
	}

	// start the timer
	go func() {
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		timeoutChannel <- true
	}()

	// wait for values
	for i := 0; i < 3; i++ {
		select {
		case data := <-dataChannel:
			log.Println("Received a value:", data)
			values = append(values, data)

		case err := <-errorsChannel:
			log.Println("Received an error:", err)
			res.WriteHeader(500)
			res.Write([]byte("One of the requests failed"))
			return

		case <-timeoutChannel:
			log.Println("The requests did not make it in time, timeout reached")
			res.WriteHeader(500)
			res.Write([]byte("Timeout reached"))
			return
		}
	}

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(common.AllResponse{
		Times: values,
	})
}

// makes a request to the sleeping server
func makeRequestToSleepServer(dataChannel chan<- int, errorChannel chan<- error) {
	res, err := http.Get("http://localhost:8080/api/sleep")

	if err != nil {
		errorChannel <- err
		return
	}

	if res.StatusCode != 200 {
		errorChannel <- errors.New("The sleep server returned an error: " + res.Status)
		return
	}

	var data common.SleepResponse
	err = json.NewDecoder(res.Body).Decode(&data)

	if err != nil {
		errorChannel <- err
		return
	}

	dataChannel <- data.Time
}
