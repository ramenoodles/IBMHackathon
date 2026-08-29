package httpapi

import (
	"encoding/json"
	"net/http"

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
