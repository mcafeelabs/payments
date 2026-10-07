package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcafeelabs/platform/pkg/sandbox"
	"github.com/mcafeelabs/platform/pkg/servicekit"
)

func TestChargeForwardsBaggage(t *testing.T) {
	identity := httptest.NewServer(sandbox.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"service": "identity", "requestSandbox": sandbox.FromContext(r.Context())})
	})))
	defer identity.Close()

	t.Setenv("SERVICE_NAME", "payments")
	t.Setenv("SANDBOX_NAME", "demo")
	k, err := servicekit.New()
	if err != nil {
		t.Fatal(err)
	}
	var cfg Config
	cfg.Identity.URL = identity.URL
	cfg.Currency = "USD"
	cfg.FeatureFlags.RefundsV2 = true
	handler(k, cfg)
	srv := httptest.NewServer(k.Handler())
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/charge", nil)
	req.Header.Set("baggage", "sandbox=demo")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&got)
	if got["sandbox"] != "demo" || got["refundsV2"] != true {
		t.Fatalf("got %v", got)
	}
	if id := got["identity"].(map[string]any); id["requestSandbox"] != "demo" {
		t.Fatalf("identity did not see the sandbox: %v", id)
	}
}
