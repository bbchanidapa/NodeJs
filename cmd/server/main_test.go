package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHello(t *testing.T) {
	// จำลอง request เข้ามาที่ path /
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// เก็บคำตอบไว้ในตัวแปร rr
	rr := httptest.NewRecorder()

	hello(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("อยากได้ status 200 แต่ได้ %d", rr.Code)
	}
	if rr.Body.String() != "Hello World!" {
		t.Errorf("อยากได้ Hello World! แต่ได้ %q", rr.Body.String())
	}
}

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("อยากได้ status 200 แต่ได้ %d", rr.Code)
	}
}
