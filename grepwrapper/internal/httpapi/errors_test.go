package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"grepwrapper/internal/service"
)

func TestErrorResponsesUseJSONContract(t *testing.T) {
	lookupService, err := service.New(t.TempDir(), "rg", nil, false)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	server := New(lookupService)
	tests := []struct {
		name    string
		path    string
		body    string
		status  int
		code    string
		message string
	}{
		{
			name:    "invalid lookup JSON",
			path:    "/v1/symbols/lookup",
			body:    "{",
			status:  http.StatusBadRequest,
			code:    errorCodeInvalidJSON,
			message: "invalid JSON",
		},
		{
			name:    "missing explain name",
			path:    "/v1/symbols/explain",
			body:    `{ "question": "what does it do?" }`,
			status:  http.StatusBadRequest,
			code:    errorCodeMissingName,
			message: "name is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", test.path, strings.NewReader(test.body))
			response := httptest.NewRecorder()

			server.Handler().ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content type = %q, want application/json", got)
			}

			var body errorResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if body.Error.Code != test.code || body.Error.Message != test.message {
				t.Fatalf("error = %#v, want code %q and message %q", body.Error, test.code, test.message)
			}
		})
	}
}

func TestLookupReturnsJSONMatches(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "sample.go"), []byte("package sample\n\nfunc Parse() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	service, err := service.New(root, "rg", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("POST", "/v1/symbols/lookup", strings.NewReader(`{"name":"Parse","language":"go"}`))
	response := httptest.NewRecorder()

	New(service).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	var body struct {
		Matches []struct {
			Path string `json:"path"`
			Line int    `json:"line"`
		} `json:"matches"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Matches) != 1 || body.Matches[0].Path != "sample.go" || body.Matches[0].Line != 3 {
		t.Fatalf("matches = %#v", body.Matches)
	}
}

func TestExplainUnavailableUsesStructuredError(t *testing.T) {
	service, err := service.New(t.TempDir(), "rg", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("POST", "/v1/symbols/explain", strings.NewReader(`{"name":"Parse"}`))
	response := httptest.NewRecorder()

	New(service).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != errorCodeLLMUnavailable {
		t.Fatalf("error code = %q", body.Error.Code)
	}
}
