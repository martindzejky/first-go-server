package requests

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/martindzejky/first-go-server/internal/responses"
	"log"
	"net/http"
)

// makes a request to the sleeping server
func MakeRequestToSleepServer(ctx context.Context, dataChannel chan<- int, errorChannel chan<- error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/api/sleep", nil)

	if checkFailed(err, errorChannel) {
		return
	}

	res, err := http.DefaultClient.Do(req)

	if checkFailed(err, errorChannel) {
		return
	}

	if res.StatusCode != 200 {
		checkFailed(errors.New("The sleep server returned an error: "+res.Status), errorChannel)
		return
	}

	var data responses.SleepResponse
	err = json.NewDecoder(res.Body).Decode(&data)

	if checkFailed(err, errorChannel) {
		return
	}

	select {
	case dataChannel <- data.Time:
	default:
		log.Println("Failed to write to dataChannel, oh well :shrug:")
	}
}

func checkFailed(err error, errorChannel chan<- error) bool {
	if err == nil {
		return false
	}

	select {
	case errorChannel <- err:
	default:
		log.Println("Failed to write to the errorChannel")
	}

	return true
}
