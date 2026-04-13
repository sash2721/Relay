package errors

import (
	"encoding/json"
	"net/http"
)

type ConflictError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	InnerError error  `json:"error"`
}

// Implement the error interface
func (e *ConflictError) Error() string {
	if e.InnerError != nil {
		return e.Message + ": " + e.InnerError.Error()
	}
	return e.Message
}

func NewConflictError(message string, err error) ([]byte, *ConflictError) {
	customError := &ConflictError{
		Code:       http.StatusConflict,
		Message:    message,
		InnerError: err, // Updated field name
	}

	jsonData, marshalErr := json.Marshal(customError)
	if marshalErr != nil {
		return []byte(`{"code":500,"message":"Internal Server Error","error":"Error while marshaling the internal server error"}`), nil
	}

	return jsonData, customError
}
