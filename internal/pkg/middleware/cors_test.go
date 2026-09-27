package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddlewareSetsHeadersAndCallsNext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	nextCalled := false
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/resource", func(c *gin.Context) {
		nextCalled = true
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/resource", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
	assertCORSHeaders(t, recorder)
}

func TestCORSMiddlewareHandlesPreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	nextCalled := false
	router := gin.New()
	router.Use(CORSMiddleware())
	router.OPTIONS("/resource", func(c *gin.Context) {
		nextCalled = true
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if nextCalled {
		t.Fatal("expected next handler not to be called")
	}
	assertCORSHeaders(t, recorder)
}

func assertCORSHeaders(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Origin, Content-Type, Accept, Authorization",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
	}

	for header, expectedValue := range expectedHeaders {
		if value := recorder.Header().Get(header); value != expectedValue {
			t.Errorf("expected %s header %q, got %q", header, expectedValue, value)
		}
	}
}
