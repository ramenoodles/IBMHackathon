package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"grepwrapper/internal/service"
)

type Server struct {
	service *service.Service
}

func New(service *service.Service) *Server {
	return &Server{
		service: service,
	}
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/symbols/lookup", server.lookup)

	mux.HandleFunc("POST /v1/symbols/explain", server.explain)

	return mux
}

func (server *Server) lookup(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request struct {
		Name     string `json:"name"`
		Language string `json:"language"`
		Limit    int    `json:"limit"`
		Context  bool   `json:"context"`
		Before   int    `json:"before"`
		After    int    `json:"after"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, errorCodeInvalidJSON, "invalid JSON")
		return
	}

	if request.Name == "" {
		writeError(w, http.StatusBadRequest, errorCodeMissingName, "name is required")
		return
	}

	results, err := server.service.Lookup(
		r.Context(),
		service.LookupRequest{
			Name:     request.Name,
			Language: request.Language,
			Limit:    request.Limit,
			Context:  request.Context,
			Before:   request.Before,
			After:    request.After,
		},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, errorCodeLookupFailed, err.Error())
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(struct {
		Matches []service.LookupResult `json:"matches"`
	}{
		Matches: results,
	}); err != nil {
		return
	}
}

func (server *Server) explain(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request struct {
		Name     string `json:"name"`
		Language string `json:"language"`
		Question string `json:"question"`
		Before   int    `json:"before"`
		After    int    `json:"after"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, errorCodeInvalidJSON, "invalid JSON")
		return
	}

	if strings.TrimSpace(request.Name) == "" {
		writeError(w, http.StatusBadRequest, errorCodeMissingName, "name is required")
		return
	}

	result, err := server.service.Explain(
		r.Context(),
		service.ExplainRequest{
			Name:     request.Name,
			Language: request.Language,
			Question: request.Question,
			Before:   request.Before,
			After:    request.After,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrLLMUnavailable):
			writeError(w, http.StatusServiceUnavailable, errorCodeLLMUnavailable, err.Error())

		case errors.Is(err, service.ErrSymbolNotFound):
			writeError(w, http.StatusNotFound, errorCodeSymbolNotFound, err.Error())

		default:
			writeError(w, http.StatusInternalServerError, errorCodeInternal, err.Error())
		}

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		return
	}
}
