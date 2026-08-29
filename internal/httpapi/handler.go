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
		http.Error(
			w,
			"invalid JSON",
			http.StatusBadRequest,
		)
		return
	}

	if request.Name == "" {
		http.Error(
			w,
			"name is required",
			http.StatusBadRequest,
		)
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
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
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
		http.Error(
			w,
			"invalid JSON",
			http.StatusBadRequest,
		)
		return
	}

	if strings.TrimSpace(request.Name) == "" {
		http.Error(
			w,
			"name is required",
			http.StatusBadRequest,
		)
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
			http.Error(
				w,
				err.Error(),
				http.StatusServiceUnavailable,
			)

		case errors.Is(err, service.ErrSymbolNotFound):
			http.Error(
				w,
				err.Error(),
				http.StatusNotFound,
			)

		default:
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
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
