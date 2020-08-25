package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/martindzejky/first-go-server/cmd/concurrent-server/requests"
	"github.com/martindzejky/first-go-server/cmd/concurrent-server/validate"
	"github.com/martindzejky/first-go-server/internal/http-utils"
	"github.com/martindzejky/first-go-server/internal/responses"
)

func main() {
	port := "8081"
	mux := http.NewServeMux()

	log.Println("Registering request handlers")
	mux.HandleFunc("/api/all", apiAllRouteHandler)
	mux.HandleFunc("/api/first", apiFirstRouteHandler)
	mux.HandleFunc("/api/within-timeout", apiWithinTimeoutRouteHandler)
	mux.HandleFunc("/api/smart", apiSmartRouteHandler)

	httpUtils.RunServer(port, mux)
}

// handles the /api/all route - it waits for and returns all 3 responses
func apiAllRouteHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request for 'all'")

	timeout, err := validate.GetAndValidateTimeout(req.URL.Query())
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
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

// handles the /api/first route, it returns the first successful response
func apiFirstRouteHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request for 'first'")

	timeout, err := validate.GetAndValidateTimeout(req.URL.Query())
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
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

// handles the /api/within-timeout route - it returns all responses within the timeout
func apiWithinTimeoutRouteHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request for 'within-validate'")

	timeout, err := validate.GetAndValidateTimeout(req.URL.Query())
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// make a new context
	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	dataChannel := make(chan int, 3)
	values := make([]int, 0)

	// make the 3 requests
	for i := 0; i < 3; i++ {
		log.Println("Making request n.", i)
		go requests.MakeRequestToSleepServer(ctx, dataChannel, nil)
	}

	doneWaiting := false

	// wait for values
	for !doneWaiting {
		select {
		case data := <-dataChannel:
			log.Println("Received a value:", data)
			values = append(values, data)

			if len(values) >= 3 {
				doneWaiting = true
			}

		case <-ctx.Done():
			doneWaiting = true
		}
	}

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(responses.AllResponse{
		Times: values,
	})
}

// handles the /api/smart route - first request made in 200ms, then 2 more if necessary
func apiSmartRouteHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Received a request for 'smart'")

	timeout, err := validate.GetAndValidateTimeout(req.URL.Query())
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// make context for the request
	ctx, cancel := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	dataChannel := make(chan int, 1)
	errorsChannel := make(chan error, 3)

	go func() {
		// make the first request
		log.Println("Making the first request")
		go requests.MakeRequestToSleepServer(ctx, dataChannel, errorsChannel)

		// make the other 2 requests after 200ms or after the first request fails
		select {
		case <-errorsChannel:
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
		}

		// if the request has been canceled/finished in the meantime,
		// do not make the requests
		if err := ctx.Err(); err != nil {
			log.Println("Not making additional requests because context error is:", err)
			return
		}

		log.Println("Making additional 2 requests")
		for i := 0; i < 2; i++ {
			go requests.MakeRequestToSleepServer(ctx, dataChannel, errorsChannel)
		}
	}()

	// wait for either the timeout or the result
	var result int
	select {
	case result = <-dataChannel:
		log.Println("Received the result:", result)

	case <-ctx.Done():
		err := ctx.Err()
		log.Println("The request failed:", err)
		http.Error(res, "The request failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(responses.AllResponse{
		Times: []int{result},
	})
}
