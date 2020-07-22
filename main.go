package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type SleepResponse struct {
    Time int
}

func main() {
	port := "8080"

	http.HandleFunc("/sleep", sleepHandler)

	fmt.Println("Server listening on port", port)
	http.ListenAndServe(":"+port, nil)
}

func sleepHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("Received a request")

    query := request.URL.Query()
    min := getQueryIntValue(query, "min", 1000)
    max := getQueryIntValue(query, "max", 4000)

    // TODO: this is deterministic
    waitTime := rand.Intn(max - min) + min

    fmt.Println("Sleeping for", waitTime, "ms")
    time.Sleep(time.Duration(waitTime) * time.Millisecond)

    writer.Header().Set("Content-Type", "application/json")
    json.NewEncoder(writer).Encode(SleepResponse{
        Time: waitTime,
    })
}

func getQueryIntValue(query url.Values, name string, defaultValue int) int {
    stringValue := query.Get(name)

    if stringValue == "" {
        return defaultValue
    }

    value, err := strconv.Atoi(stringValue)

    if err != nil {
        return defaultValue
    }

    return value
}
