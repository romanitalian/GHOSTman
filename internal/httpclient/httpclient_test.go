package httpclient

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRequest(t *testing.T) {
	rq, err := NewRequest("POST", "http://example.com", "body", "X-Test: 123\nContent-Type: text/plain")
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}
	if rq.Method != "POST" || rq.URL.String() != "http://example.com" {
		t.Errorf("unexpected method or url: %s %s", rq.Method, rq.URL)
	}
	if rq.Header.Get("X-Test") != "123" || rq.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("unexpected headers: %v", rq.Header)
	}
}

func TestSendRequest(t *testing.T) {
	// Мок-сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/err" {
			w.WriteHeader(500)
			w.Write([]byte("fail"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	rq, _ := NewRequest("GET", ts.URL, "", "")
	status, body, isErr, err := SendRequest(rq)
	if err != nil || isErr {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(status, "200") || !strings.Contains(body, "ok") {
		t.Errorf("unexpected status/body: %s %s", status, body)
	}

	rq2, _ := NewRequest("GET", ts.URL+"/err", "", "")
	status2, body2, isErr2, err2 := SendRequest(rq2)
	if err2 != nil || !isErr2 || !strings.Contains(status2, "500") || !strings.Contains(body2, "fail") {
		t.Errorf("unexpected error or response: %v %v %s %s", isErr2, err2, status2, body2)
	}
}
