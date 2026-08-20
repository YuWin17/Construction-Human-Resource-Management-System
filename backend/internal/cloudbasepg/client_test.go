package cloudbasepg

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientFetchAndReplaceTable(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer server-only-key" {
			t.Fatalf("authorization header = %q", got)
		}
		methods = append(methods, r.Method)
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("select") != "*" {
				t.Fatalf("select query = %q", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[{"id":"talent-1","name":"张三"}]`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPost:
			var row map[string]any
			if err := json.NewDecoder(r.Body).Decode(&row); err != nil {
				t.Fatalf("decode row: %v", err)
			}
			if row["id"] != "talent-2" {
				t.Fatalf("unexpected row: %#v", row)
			}
			w.WriteHeader(http.StatusCreated)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client, err := NewWithBaseURL(server.URL, "server-only-key", server.Client())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	rows, err := client.FetchTable(context.Background(), "talents")
	if err != nil {
		t.Fatalf("fetch table: %v", err)
	}
	if len(rows) != 1 || rows[0]["name"] != "张三" {
		t.Fatalf("fetched rows: %#v", rows)
	}
	if err := client.ReplaceTable(context.Background(), "talents", []map[string]any{{"id": "talent-2"}}); err != nil {
		t.Fatalf("replace table: %v", err)
	}
	if got, want := methods, []string{http.MethodGet, http.MethodDelete, http.MethodPost}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("methods = %#v, want %#v", got, want)
	}
}
