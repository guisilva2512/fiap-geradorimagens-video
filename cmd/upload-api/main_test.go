package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerProtectsUploadRoutes(t *testing.T) {
	router := server(nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/uploads", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}
