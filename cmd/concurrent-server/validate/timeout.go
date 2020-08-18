package validate

import (
	"errors"
	httpUtils "github.com/martindzejky/first-go-server/internal/http-utils"
	"log"
	"net/url"
)

func GetAndValidateTimeout(query url.Values) (int, error) {
	timeout := httpUtils.GetQueryIntValue(query, "timeout", 1000)
	if timeout < 100 || timeout > 5000 {
		log.Println("Invalid timeout received:", timeout)
		return 0, errors.New("incorrect value for timeout specified, it must be 100 < timeout < 5000")
	}

	return timeout, nil
}
