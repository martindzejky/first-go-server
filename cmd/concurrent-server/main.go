package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/martindzejky/first-go-server/internal/http-utils"
	"github.com/martindzejky/first-go-server/internal/responses"
)

func main() {
	port := "8081"

	// register listeners
	log.Println("Registering listeners")
	http.HandleFunc("/api/all", apiAllRoute)

	// start the server
	log.Println("Starting the server on port", port)
	log.Fatalln(http.ListenAndServe(":"+port, nil))
}

// handles the /api/all route - it waits for and returns all 3 responses
func apiAllRoute(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request for 'all'")

	// validate timeout
	timeout := httpUtils.GetQueryIntValue(req.URL.Query(), "timeout", 1000)
	if timeout < 100 || timeout > 5000 {
		log.Println("Invalid timeout received:", timeout)
		http.Error(res, "Incorrect value for timeout specified, it must be 100 < timeout < 5000", http.StatusBadRequest)
		return
	}

	// make a new context
	ctx, _ := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Millisecond)

	dataChannel := make(chan int, 3)
	errorsChannel := make(chan error, 3)
	values := make([]int, 0)

	// make the 3 requests
	for i := 0; i < 3; i++ {
		log.Println("Making request n.", i)
		go makeRequestToSleepServer(dataChannel, errorsChannel)
	}

	// wait for values
	for i := 0; i < 3; i++ {
		select {
		case data := <-dataChannel:
			log.Println("Received a value:", data)
			values = append(values, data)

		case <-errorsChannel:
			log.Println("Received an error")

		case <-ctx.Done():
			log.Println("The requests did not make it in time or it was canceled")
			http.Error(res, "Timeout reached", http.StatusInternalServerError)
			return
		}
	}

	if len(values) == 0 {
		log.Println("None of the requests was successful")
		http.Error(res, "All requests failed", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(responses.AllResponse{
		Times: values,
	})
}

// makes a request to the sleeping server
func makeRequestToSleepServer(dataChannel chan<- int, errorChannel chan<- error) {
	res, err := http.Get("http://localhost:8080/api/sleep")

	if err != nil {
		select {
		case errorChannel <- err:
		default:
			log.Println("Failed to write to errorChannel")
		}

		return
	}

	if res.StatusCode != 200 {
		select {
		case errorChannel <- errors.New("The sleep server returned an error: " + res.Status):
		default:
			log.Println("Failed to write to errorChannel")
		}

		return
	}

	var data responses.SleepResponse
	err = json.NewDecoder(res.Body).Decode(&data)

	if err != nil {
		select {
		case errorChannel <- err:
		default:
			log.Println("Failed to write to errorChannel")
		}

		return
	}

	select {
	case dataChannel <- data.Time:
	default:
		log.Println("Failed to write to dataChannel, oh well")
	}
}
