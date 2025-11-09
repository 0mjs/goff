package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0mjs/goff"
)

func main() {
	var (
		yamlPath = flag.String("yaml", "../admin/flags.yaml", "Path to flags YAML file")
		port     = flag.String("port", "8080", "Server port")
	)
	flag.Parse()

	// Initialize goff client with With... methods
	client, err := goff.New(
		goff.WithFile(*yamlPath),
		goff.WithAutoReload(2*time.Second),
		goff.WithHooks(goff.Hooks{
			AfterEval: func(flag, variant string, reason goff.Reason) {
				// Log flag evaluations
				log.Printf("[FLAG] flag=%s variant=%s reason=%v", flag, variant, reason)
			},
		}),
	)
	if err != nil {
		log.Fatalf("Failed to initialize goff: %v", err)
	}
	defer client.Close()

	// Endpoint 1: Checkout with different behaviors based on flags
	http.HandleFunc("/checkout", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user")
		if userID == "" {
			userID = "anonymous"
		}

		plan := r.URL.Query().Get("plan")
		if plan == "" {
			plan = "basic"
		}

		ctx := goff.Context{
			Key: fmt.Sprintf("user:%s", userID),
			Attrs: map[string]any{
				"plan": plan,
			},
		}

		newCheckout := client.Boolean("new_checkout", ctx, false)
		theme := client.String("checkout_theme", ctx, "default")

		// Different outcomes based on flags
		if newCheckout {
			log.Printf("[CHECKOUT] User %s using NEW checkout flow with theme: %s", userID, theme)
		} else {
			log.Printf("[CHECKOUT] User %s using OLD checkout flow", userID)
		}

		response := map[string]any{
			"user_id":      userID,
			"new_checkout": newCheckout,
			"theme":        theme,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Endpoint 2: Logging behavior based on flags
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user")
		if userID == "" {
			userID = "anonymous"
		}

		ctx := goff.Context{
			Key: fmt.Sprintf("user:%s", userID),
			Attrs: map[string]any{},
		}

		enableLogging := client.Boolean("enable_logging", ctx, true)
		logLevel := client.String("log_level", ctx, "info")

		// Different logging outcomes based on flags
		if enableLogging {
			switch logLevel {
			case "debug":
				log.Printf("[DEBUG] Fetching users for user: %s", userID)
			case "info":
				log.Printf("[INFO] Fetching users")
			case "warn":
				log.Printf("[WARN] Fetching users")
			case "error":
				// Only log errors
			default:
				log.Printf("[%s] Fetching users", logLevel)
			}
		} else {
			// No logging when disabled
		}

		response := map[string]any{
			"users": []map[string]any{
				{"id": "1", "name": "Alice"},
				{"id": "2", "name": "Bob"},
			},
			"logging_enabled": enableLogging,
			"log_level":        logLevel,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Endpoint 3: Feature-based routing
	http.HandleFunc("/api/features", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user")
		if userID == "" {
			userID = "anonymous"
		}

		ctx := goff.Context{
			Key: fmt.Sprintf("user:%s", userID),
			Attrs: map[string]any{},
		}

		newCheckout := client.Boolean("new_checkout", ctx, false)
		enableLogging := client.Boolean("enable_logging", ctx, true)

		features := map[string]any{
			"new_checkout":  newCheckout,
			"enable_logging": enableLogging,
		}

		if newCheckout {
			log.Printf("[FEATURES] User %s has access to new checkout", userID)
		}
		if enableLogging {
			log.Printf("[FEATURES] Logging enabled for user %s", userID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(features)
	})

	// Test endpoint: Demonstrate checkout_theme with attributes
	http.HandleFunc("/test/theme", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user")
		if userID == "" {
			userID = "test-user"
		}

		themeAttr := r.URL.Query().Get("theme")
		count := r.URL.Query().Get("count")
		if count == "" {
			count = "1"
		}
		
		// Create context WITH theme attribute
		ctxWithTheme := goff.Context{
			Key: fmt.Sprintf("user:%s", userID),
			Attrs: map[string]any{
				"theme": themeAttr,
			},
		}

		// Create context WITHOUT theme attribute (for comparison)
		ctxWithoutTheme := goff.Context{
			Key: fmt.Sprintf("user:%s", userID),
			Attrs: map[string]any{},
		}

		// Evaluate multiple times to show sticky bucketing
		var resultsWithAttr []string
		var resultsWithoutAttr []string
		for i := 0; i < 5; i++ {
			resultsWithAttr = append(resultsWithAttr, client.String("checkout_theme", ctxWithTheme, "default"))
			resultsWithoutAttr = append(resultsWithoutAttr, client.String("checkout_theme", ctxWithoutTheme, "default"))
		}

		// Determine which rule matched
		ruleMatched := ""
		if themeAttr == "dark" {
			ruleMatched = "Rule matched: theme='dark' → variants: {black: 100%} → ALWAYS returns 'black'"
		} else if themeAttr != "" {
			ruleMatched = fmt.Sprintf("No rule matched for theme='%s' → using default variants (red: 40%%, blue: 30%%, green: 30%%)", themeAttr)
		} else {
			ruleMatched = "No theme attribute provided → using default variants (red: 40%, blue: 30%, green: 30%)"
		}

		// Check if results are consistent (sticky bucketing)
		allSameWith := true
		allSameWithout := true
		for i := 1; i < len(resultsWithAttr); i++ {
			if resultsWithAttr[i] != resultsWithAttr[0] {
				allSameWith = false
			}
			if resultsWithoutAttr[i] != resultsWithoutAttr[0] {
				allSameWithout = false
			}
		}

		response := map[string]any{
			"description": "Testing checkout_theme flag with theme attribute",
			"note":        "Same user = same result (sticky bucketing). Different users = distribution across users.",
			"user_id":     userID,
			"input": map[string]any{
				"theme_attribute": themeAttr,
				"evaluations":     5,
			},
			"evaluation": map[string]any{
				"with_theme_attr": map[string]any{
					"context": map[string]any{
						"key":   fmt.Sprintf("user:%s", userID),
						"attrs": map[string]any{"theme": themeAttr},
					},
					"results":      resultsWithAttr,
					"is_consistent": allSameWith,
					"explanation":   "Same user + same attributes = same result every time (sticky bucketing)",
					"rule_info":     ruleMatched,
				},
				"without_theme_attr": map[string]any{
					"context": map[string]any{
						"key":   fmt.Sprintf("user:%s", userID),
						"attrs": map[string]any{},
					},
					"results":      resultsWithoutAttr,
					"is_consistent": allSameWithout,
					"explanation":   "Same user = same result every time (sticky bucketing)",
					"rule_info":     "No theme attribute → default variants (red: 40%, blue: 30%, green: 30%)",
				},
			},
			"flag_config": map[string]any{
				"default_variants": map[string]int{
					"red":   40,
					"blue":  30,
					"green": 30,
				},
				"rules": []map[string]any{
					{
						"when": "theme == 'dark'",
						"then": "black: 100% (deterministic - always returns black)",
					},
				},
			},
			"how_it_works": map[string]any{
				"sticky_bucketing": "Same user ID always gets the same variant (deterministic hash)",
				"rules":            "Rules are evaluated first - if matched, use rule variants (deterministic)",
				"variants":         "Percentage variants are distributed across DIFFERENT users, not requests",
				"example":          "Try: /test/theme?user=alice (5 times) → same result. Try: /test/theme?user=bob → may get different result.",
			},
			"examples": []string{
				"/test/theme?user=alice&theme=dark    → black (rule matched, deterministic)",
				"/test/theme?user=bob&theme=light     → red/blue/green (sticky per user)",
				"/test/theme?user=charlie              → red/blue/green (sticky per user)",
				"Try same user 5 times → always same result",
				"Try different users → see distribution across users",
			},
		}

		log.Printf("[TEST/THEME] User=%s theme_attr=%q results_with=%v results_without=%v", 
			userID, themeAttr, resultsWithAttr, resultsWithoutAttr)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("App server starting on :%s", *port)
	log.Printf("Flags YAML: %s", *yamlPath)
	log.Println("Try:")
	log.Println("  http://localhost:" + *port + "/checkout?user=123&plan=pro")
	log.Println("  http://localhost:" + *port + "/api/users?user=456")
	log.Println("  http://localhost:" + *port + "/api/features?user=789")
	log.Println("  http://localhost:" + *port + "/test/theme?user=alice&theme=dark")
	log.Println("  http://localhost:" + *port + "/test/theme?user=bob&theme=light")

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

