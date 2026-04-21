package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestHello(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
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

func TestUUID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/uuid", nil)
	rr := httptest.NewRecorder()

	newUUID(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("อยากได้ status 200 แต่ได้ %d", rr.Code)
	}
	body := rr.Body.String()
	if _, err := uuid.Parse(body); err != nil {
		t.Errorf("คำตอบไม่ใช่ UUID ที่อ่านได้: %q err=%v", body, err)
	}
}

func TestCreateUserWhenPostgresNotConfigured(t *testing.T) {
	body := []byte(`{"firstname":"Bee","lastname":"Coder"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := createUserHandler(nil)
	handler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("อยากได้ status 503 แต่ได้ %d", rr.Code)
	}
}

func TestListUsersWhenPostgresNotConfigured(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler := listUsersHandler(nil)
	handler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("อยากได้ status 503 แต่ได้ %d", rr.Code)
	}
}

func TestGetUserByIDWhenPostgresNotConfigured(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/6d4cbf4e-94f5-4e0f-af4f-b5f96ce267b9", nil)
	rr := httptest.NewRecorder()

	handler := getUserByIDHandler(nil)
	handler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("อยากได้ status 503 แต่ได้ %d", rr.Code)
	}
}

func TestUpdateUserByIDWhenPostgresNotConfigured(t *testing.T) {
	body := []byte(`{"firstname":"Bee","lastname":"Coder"}`)
	req := httptest.NewRequest(http.MethodPut, "/users/6d4cbf4e-94f5-4e0f-af4f-b5f96ce267b9", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := updateUserByIDHandler(nil)
	handler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("อยากได้ status 503 แต่ได้ %d", rr.Code)
	}
}

func TestPatchUserByIDWhenPostgresNotConfigured(t *testing.T) {
	body := []byte(`{"firstname":"Bee","lastname":"Coder"}`)
	req := httptest.NewRequest(http.MethodPatch, "/users/6d4cbf4e-94f5-4e0f-af4f-b5f96ce267b9", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := userByIDHandler(nil)
	handler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("อยากได้ status 503 แต่ได้ %d", rr.Code)
	}
}

func TestPutUsersWithoutIDReturnsBadRequest(t *testing.T) {
	body := []byte(`{"firstname":"Bee","lastname":"Coder"}`)
	req := httptest.NewRequest(http.MethodPut, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := usersHandler(nil)
	handler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("อยากได้ status 400 แต่ได้ %d", rr.Code)
	}
}
