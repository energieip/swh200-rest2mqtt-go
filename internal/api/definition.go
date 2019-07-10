package api

import (
	"encoding/json"
	"net/http"

	"github.com/romana/rlog"
)

const (
	APIErrorDeviceNotFound = 1
	APIErrorBodyParsing    = 2
	APIErrorDatabase       = 3
	APIErrorInvalidValue   = 4
	APIErrorUnauthorized   = 5
	APIErrorExpiredToken   = 6

	TokenName = "EiPAccessToken"
)

//APIError Message error code
type APIError struct {
	Code    int    `json:"code"` //errorCode
	Message string `json:"message"`
}

type API struct {
	EventsToBackend chan map[string]interface{}
	certificate     string
	keyfile         string
	apiIP           string
	apiPort         string
	apiPassword     string
	browsingFolder  string
	dataPath        string
}

type APIInfo struct {
	Versions []string `json:"versions"`
}

type APIFunctions struct {
	Functions []string `json:"functions"`
}

type networkError struct {
	s string
}

func (e *networkError) Error() string {
	return e.s
}

// NewError raise an error
func NewError(text string) error {
	return &networkError{text}
}

func (api *API) sendError(w http.ResponseWriter, errorCode int, message string, httpStatus int) {
	errCode := APIError{
		Code:    errorCode,
		Message: message,
	}

	inrec, _ := json.MarshalIndent(errCode, "", "  ")
	rlog.Error(errCode.Message)
	http.Error(w, string(inrec), httpStatus)
}
