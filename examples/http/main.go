package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0mjs/goff"
)

func main() {
	// Initialize goff client
	client, err := goff.New(
		goff.WithFile("../../testdata/flags.yaml"),
		goff.WithAutoReload(5*time.Second),
		goff.WithHooks(goff.Hooks{
			AfterEval: func(flag, variant string, reason goff.Reason) {
				// Log flag evaluations (in production, send to metrics system)
				log.Printf("flag=%s variant=%s reason=%v", flag, variant, reason)
			},
		}),
	)
	if err != nil {
		log.Fatalf("Failed to initialize goff: %v", err)
	}
	defer client.Close()

	// HTTP handler
	http.HandleFunc("/checkout", func(w http.ResponseWriter, r *http.Request) {
		// Extract user context from request
		userID := r.URL.Query().Get("user")
		if userID == "" {
			userID = "anonymous"
		}

		plan := r.URL.Query().Get("plan")
		if plan == "" {
			plan = "basic"
		}

		// Create evaluation context
		ctx := goff.Context{
			Key: fmt.Sprintf("user:%s", userID),
			Attrs: map[string]interface{}{
				"plan": plan,
			},
		}

		// Evaluate flags
		newCheckout := client.Bool("new_checkout", ctx, false)
		theme := client.String("checkout_theme", ctx, "default")

		// Build response
		response := map[string]interface{}{
			"user_id":      userID,
			"new_checkout": newCheckout,
			"theme":        theme,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Server starting on :8080")
	log.Println("Try: http://localhost:8080/checkout?user=123&plan=pro")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

