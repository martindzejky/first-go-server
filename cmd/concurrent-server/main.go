package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/martindzejky/first-go-server/cmd/concurrent-server/requests"
	osSignals "github.com/martindzejky/first-go-server/internal/os-signals"
	"log"
	"net/http"
	"time"

	"github.com/martindzejky/first-go-server/internal/http-utils"
	"github.com/martindzejky/first-go-server/internal/responses"
)

func main() {
	port := "8081"
	router := mux.NewRouter()

	server := httpUtils.MakeServer(port, router)
	signals := osSignals.MakeChannelWithInterruptSignal()

	log.Println("Registering request handlers")
	router.HandleFunc("/api/all", apiAllRouteHandler)
	router.HandleFunc("/api/first", apiFirstRouteHandler)

	go func() {
		log.Println("Starting the server")
		err := server.ListenAndServe()

		if err != nil {
			log.Fatalln(err)
		}
	}()

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

// handles the /api/all route - it waits for and returns all 3 responses
func apiAllRouteHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request for 'all'")

	// validate timeout
	timeout := httpUtils.GetQueryIntValue(req.URL.Query(), "timeout", 1000)
	if timeout < 100 || timeout > 5000 {
		log.Println("Invalid timeout received:", timeout)
		http.Error(res, "Incorrect value for timeout specified, it must be 100 < timeout < 5000", http.StatusBadRequest)
		return
	}

	// make a new context
	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	dataChannel := make(chan int, 3)
	errorsChannel := make(chan error, 3)
	values := make([]int, 0)

	// make the 3 requests
	for i := 0; i < 3; i++ {
		log.Println("Making request n.", i)
		go requests.MakeRequestToSleepServer(ctx, dataChannel, errorsChannel)
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
			log.Println("The request failed:", ctx.Err())
			http.Error(res, "The request failed: "+ctx.Err().Error(), http.StatusInternalServerError)
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

func apiFirstRouteHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request for 'first'")

	// validate timeout
	timeout := httpUtils.GetQueryIntValue(req.URL.Query(), "timeout", 1000)
	if timeout < 100 || timeout > 5000 {
		log.Println("Invalid timeout received:", timeout)
		http.Error(res, "Incorrect value for timeout specified, it must be 100 < timeout < 5000", http.StatusBadRequest)
		return
	}

	// make a new context
	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	dataChannel := make(chan int, 1)
	errorsChannel := make(chan error, 3)

	// make the 3 requests
	for i := 0; i < 3; i++ {
		log.Println("Making request n.", i)
		go requests.MakeRequestToSleepServer(ctx, dataChannel, errorsChannel)
	}

	var (
		result         int
		receivedResult = false
		failedRequests int
	)

	// wait for values
	for !receivedResult {
		select {
		case result = <-dataChannel:
			receivedResult = true
			log.Println("Received a value:", result)

		case <-errorsChannel:
			failedRequests++
			log.Println("Received an error")

			if failedRequests >= 3 {
				log.Println("None of the requests was successful")
				http.Error(res, "All requests failed", http.StatusInternalServerError)
				return
			}

		case <-ctx.Done():
			log.Println("The request failed:", ctx.Err())
			http.Error(res, "The request failed: "+ctx.Err().Error(), http.StatusInternalServerError)
			return
		}
	}

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(responses.AllResponse{
		Times: []int{result},
	})
}
