package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetPolkaAPIKey(headers http.Header) (string, error) {
	polkaKey := headers.Get("Authorization")
	splittedValue := strings.Fields(polkaKey)

	if len(splittedValue) == 0 {
		return "", fmt.Errorf("api key is missing")
	}

	if splittedValue[0] != "ApiKey" || len(splittedValue) != 2 {
		return "", fmt.Errorf("malformed api key")
	}

	return splittedValue[1], nil
}
