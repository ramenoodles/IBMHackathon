package httpapi

import (
	"encoding/json"
	"net/http"
)

const (
	errorCodeInvalidJSON    = "invalid_json"
	errorCodeMissingName    = "missing_name"
	errorCodeLLMUnavailable = "llm_unavailable"
	errorCodeSymbolNotFound = "symbol_not_found"
	errorCodeLookupFailed   = "lookup_failed"
	errorCodeInternal       = "internal_error"
)

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Error: errorDetail{
			Code:    code,
			Message: message,
		},
	})
}
