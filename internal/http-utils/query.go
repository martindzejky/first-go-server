package httpUtils

import (
	"net/url"
	"strconv"
)

// get the value of a query parameter as an int
func GetQueryIntValue(query url.Values, name string, defaultValue int) int {
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
