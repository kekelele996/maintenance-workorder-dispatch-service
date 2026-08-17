package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"workorder/internal/config"
	"workorder/internal/repository"
	"workorder/internal/service"
	"workorder/internal/store"
)

func newServer() http.Handler {
	st := store.New()
	repo := repository.New(st)
	svc := service.New(repo, config.Load())
	return New(svc).Routes()
}

func TestGetMissingOrderReturns404(t *testing.T) {
	h := newServer()
	req := httptest.NewRequest(http.MethodGet, "/orders/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
}

func TestCreateAndGetOrder(t *testing.T) {
	h := newServer()
	body := `{"equipment_id":"CNC-01","title":"主轴异响","priority":3}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status=%d want 201", rec.Code)
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id := created["ID"].(string)
	getReq := httptest.NewRequest(http.MethodGet, "/orders/"+id, nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status=%d want 200", getRec.Code)
	}
}

func TestListOrders(t *testing.T) {
	h := newServer()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d want 200", rec.Code)
	}
}
