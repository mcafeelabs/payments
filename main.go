// Command payments is a demo service: /charge calls identity and reports
// which copy of each service handled the request.
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/mcafeelabs/platform/pkg/servicekit"
)

// Config is config-values.schema.yaml.
type Config struct {
	Identity struct {
		URL string `json:"url"`
	} `json:"identity"`
	Currency     string `json:"currency"`
	FeatureFlags struct {
		RefundsV2 bool `json:"refundsV2"`
	} `json:"featureFlags"`
}

func handler(k *servicekit.Kit, cfg Config) {
	k.Mux.HandleFunc("POST /charge", func(w http.ResponseWriter, r *http.Request) {
		var identity map[string]any
		if err := k.Call(r.Context(), http.MethodGet, cfg.Identity.URL+"/verify", nil, &identity); err != nil {
			k.JSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
			return
		}
		info := k.Info(r.Context())
		info["currency"] = cfg.Currency
		info["refundsV2"] = cfg.FeatureFlags.RefundsV2
		info["identity"] = identity
		k.JSON(w, http.StatusOK, info)
	})
}

func main() {
	k, err := servicekit.New()
	if err != nil {
		log.Fatal(err)
	}
	var cfg Config
	if err := k.Decode(&cfg); err != nil {
		log.Fatal(err)
	}
	handler(k, cfg)
	if err := k.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
