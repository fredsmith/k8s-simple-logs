package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHealthcheck tests the /healthcheck endpoint
func TestHealthcheck(t *testing.T) {
	// Skip if running in environment without k8s config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); os.IsNotExist(err) {
		t.Skip("Skipping test - not running in Kubernetes cluster")
	}

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/healthcheck", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "still alive", w.Body.String())
}

// TestLogsEndpointWithoutKey tests /logs endpoint when LOGKEY is not set
func TestLogsEndpointWithoutKey(t *testing.T) {
	// Skip if running in environment without k8s config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); os.IsNotExist(err) {
		t.Skip("Skipping test - not running in Kubernetes cluster")
	}

	// Ensure LOGKEY is not set
	os.Unsetenv("LOGKEY")

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logs", nil)
	router.ServeHTTP(w, req)

	// Should succeed when no LOGKEY is required
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestLogsEndpointWithValidKey tests /logs endpoint with correct key
func TestLogsEndpointWithValidKey(t *testing.T) {
	// Skip if running in environment without k8s config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); os.IsNotExist(err) {
		t.Skip("Skipping test - not running in Kubernetes cluster")
	}

	// Set LOGKEY
	os.Setenv("LOGKEY", "testkey123")
	defer os.Unsetenv("LOGKEY")

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logs?key=testkey123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestLogsEndpointWithInvalidKey tests /logs endpoint with incorrect key
func TestLogsEndpointWithInvalidKey(t *testing.T) {
	// Skip if running in environment without k8s config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); os.IsNotExist(err) {
		t.Skip("Skipping test - not running in Kubernetes cluster")
	}

	// Set LOGKEY
	os.Setenv("LOGKEY", "correctkey")
	defer os.Unsetenv("LOGKEY")

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logs?key=wrongkey", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "Key Required", w.Body.String())
}

// TestLogsEndpointWithMissingKey tests /logs endpoint when key is required but missing
func TestLogsEndpointWithMissingKey(t *testing.T) {
	// Skip if running in environment without k8s config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); os.IsNotExist(err) {
		t.Skip("Skipping test - not running in Kubernetes cluster")
	}

	// Set LOGKEY
	os.Setenv("LOGKEY", "requiredkey")
	defer os.Unsetenv("LOGKEY")

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "Key Required", w.Body.String())
}

// TestLogsEndpointWithCustomLines tests /logs endpoint with custom lines parameter
func TestLogsEndpointWithCustomLines(t *testing.T) {
	// Skip if running in environment without k8s config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); os.IsNotExist(err) {
		t.Skip("Skipping test - not running in Kubernetes cluster")
	}

	os.Unsetenv("LOGKEY")

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logs?lines=50", nil)
	router.ServeHTTP(w, req)

	// Should succeed with custom lines parameter
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestLogsEndpointWithInvalidLinesParameter tests /logs with non-numeric lines param
func TestLogsEndpointWithInvalidLinesParameter(t *testing.T) {
	// Skip if running in environment without k8s config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); os.IsNotExist(err) {
		t.Skip("Skipping test - not running in Kubernetes cluster")
	}

	os.Unsetenv("LOGKEY")

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logs?lines=invalid", nil)
	router.ServeHTTP(w, req)

	// Should fall back to default (20 lines) and still succeed
	assert.Equal(t, http.StatusOK, w.Code)
}
