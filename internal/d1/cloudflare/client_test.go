package cloudflare

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testDatabaseID = "11111111-1111-1111-1111-111111111111"

func TestQueryUsesD1API(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/accounts/account/d1/database/"+testDatabaseID+"/query" {
			t.Fatalf("unexpected request path %q", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("unexpected authorization header")
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"success":true,"result":[{"success":true,"results":[{"name":"widgets"}]}]}`))
	}))
	defer server.Close()
	client, err := New(Config{
		AccountID: "account",
		Database:  testDatabaseID,
		APIToken:  "secret",
		BaseURL:   server.URL,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	rows, err := client.Query(context.Background(), "SELECT name FROM sqlite_schema")
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if len(rows) != 1 || rows[0]["name"] != "widgets" {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestNamedDatabaseIsResolvedBeforeQuery(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		response.Header().Set("Content-Type", "application/json")
		if request.Method == http.MethodGet {
			if request.URL.Query().Get("name") != "production" {
				t.Fatalf("unexpected name query: %s", request.URL.RawQuery)
			}
			_, _ = response.Write([]byte(`{"success":true,"result":[{"uuid":"` + testDatabaseID + `","name":"production"}]}`))
			return
		}
		if !strings.Contains(request.URL.Path, testDatabaseID) {
			t.Fatalf("query did not use resolved id: %s", request.URL.Path)
		}
		_, _ = response.Write([]byte(`{"success":true,"result":[{"success":true,"results":[]}]}`))
	}))
	defer server.Close()
	client, err := New(Config{
		AccountID: "account",
		Database:  "production",
		APIToken:  "secret",
		BaseURL:   server.URL,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := client.Query(context.Background(), "SELECT 1"); err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestBatchSendsTransactionalQueryBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var body struct {
			Batch []struct {
				SQL string `json:"sql"`
			} `json:"batch"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(body.Batch) != 2 || body.Batch[1].SQL != "CREATE TABLE widgets(id INTEGER)" {
			t.Fatalf("unexpected batch: %#v", body.Batch)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"success":true,"result":[{"success":true},{"success":true}]}`))
	}))
	defer server.Close()
	client, err := New(Config{
		AccountID: "account",
		Database:  testDatabaseID,
		APIToken:  "secret",
		BaseURL:   server.URL,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if err := client.Batch(context.Background(), []string{
		"PRAGMA defer_foreign_keys = ON",
		"CREATE TABLE widgets(id INTEGER)",
	}); err != nil {
		t.Fatalf("Batch returned error: %v", err)
	}
}

func TestQueryReportsCloudflareError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.WriteHeader(http.StatusBadRequest)
		_, _ = response.Write([]byte(`{"success":false,"errors":[{"code":7500,"message":"bad query"}]}`))
	}))
	defer server.Close()
	client, err := New(Config{
		AccountID: "account",
		Database:  testDatabaseID,
		APIToken:  "secret",
		BaseURL:   server.URL,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	_, err = client.Query(context.Background(), "broken")
	if err == nil || !strings.Contains(err.Error(), "7500: bad query") {
		t.Fatalf("unexpected error: %v", err)
	}
}
