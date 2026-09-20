package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerProtectsUserListing(t *testing.T) {
	router := server(nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/users", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}
