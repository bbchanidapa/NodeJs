package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoot(t *testing.T) {
	srv := httptest.NewServer(newMux())
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var b strings.Builder
	if _, err := b.ReadFrom(res.Body); err != nil {
		t.Fatal(err)
	}
	if got := b.String(); got != "Hello World!" {
		t.Fatalf("body %q", got)
	}
}

func TestHealth(t *testing.T) {
	srv := httptest.NewServer(newMux())
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
}
