package api

import (
	"encoding/json"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"message": "Stock Analysis API with AI-powered recommendations",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"daily_recommendations": "GET /api/stock/daily-recommendations",
			"analyze_stock":         "POST /api/stock/analyze",
			"general_ai":            "POST /api/prompt",
			"health":                "GET /api/health",
		},
	}

	json.NewEncoder(w).Encode(response)
}
