package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestFrontend_ServedAtRoot(t *testing.T) {
	rec := doGet(newTestHandler(&fakeService{}), "/", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "<title>Trade for All</title>") {
		t.Fatal("page body does not look like the frontend")
	}
}

func TestFrontend_UnknownPathIsNotFound(t *testing.T) {
	rec := doGet(newTestHandler(&fakeService{}), "/no-such-page", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
