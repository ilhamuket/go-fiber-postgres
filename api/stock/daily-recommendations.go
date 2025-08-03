package stock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type StockRecommendationResponse struct {
	Status   string `json:"status"`
	Date     string `json:"date"`
	Analysis string `json:"analysis"`
	Error    string `json:"error,omitempty"`
}

func DailyRecommendations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, X-goog-api-key")
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

	currentDate := time.Now().Format("2006-01-02")

	prompt := fmt.Sprintf(`Anda adalah head trader di investment firm Jakarta dengan 15 tahun pengalaman trading saham Indonesia. Client VIP meminta daily picks untuk modal Rp 7.5 juta pada %s.

TRADING MANDATE:
- Capital: Rp 7,500,000
- Target: 4-5%% per trade
- Style: Active day/swing trading
- Timeline: 1-3 days per position

MARKET BRIEFING %s:

**IHSG STATUS**
Current level: [Estimate based on typical range]
Trend: [Bullish/Bearish/Sideways]
Key resistance: [Level]
Key support: [Level]

**SECTOR ROTATION**
Outperforming: [Which sectors leading]
Underperforming: [Weak sectors]
Foreign flow: [Net buy/sell estimate]

**TOP 4 TRADING OPPORTUNITIES**

**PICK 1: BLUE CHIP DEFENSIVE**
Stock: [Choose from BBCA, BBRI, ASII, UNVR]
Price: Rp [Realistic current estimate]
Why now: [Specific catalyst]
Technical: [Pattern, RSI, support/resistance]
Entry: Rp [range]
Target: Rp [4-5%% up]
Stop: Rp [level]
Size: Rp [amount from 7.5M]
Risk: Low-Medium

**PICK 2: GROWTH/IPO MOMENTUM**  
Stock: [CDIA, GOTO, or similar growth play]
Price: Rp [Realistic estimate]
Why now: [Growth catalyst or momentum]
Technical: [Breakout, volume, momentum indicators]
Entry: Rp [range]
Target: Rp [4-5%% up but account for volatility]
Stop: Rp [tighter due to volatility]
Size: Rp [smaller allocation due to risk]
Risk: High

**PICK 3: RECOVERY VALUE**
Stock: [Oversold quality name]
Price: Rp [Current depressed level]
Why now: [Oversold bounce opportunity]
Technical: [Reversal signals, support test]
Entry: Rp [at support]
Target: Rp [bounce target]
Stop: Rp [below support]
Size: Rp [amount]
Risk: Medium

**PICK 4: MOMENTUM BREAKOUT**
Stock: [High beta momentum play]
Price: Rp [Current price near breakout]
Why now: [Breakout setup, volume surge]
Technical: [Pattern completion, momentum]
Entry: Rp [breakout level]
Target: Rp [measured move]
Stop: Rp [below breakout]
Size: Rp [amount]
Risk: Medium-High

**PORTFOLIO ALLOCATION**
Total deployed: Rp [sum of positions]
Cash reserve: Rp [remaining]
Max single position: 30%% of capital
Correlation check: [Ensure diversification]

**EXECUTION STRATEGY**
09:00-09:30: [Opening gap analysis]
10:30-11:30: [Mid-session momentum]
13:30-15:00: [Afternoon positioning]

**RISK MANAGEMENT**
Portfolio stop: [If IHSG breaks X level]
Individual stops: [Price-based, not time-based]
Profit taking: [25%% at 3%%, 50%% at 4%%, remainder at 5%%]

Provide actionable recommendations with specific stock names, realistic prices, and clear entry/exit levels. Focus on liquid Indonesian stocks suitable for Rp 7.5M capital deployment.`, currentDate, currentDate)

	response, err := callGemini2API(prompt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(StockRecommendationResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(StockRecommendationResponse{
		Status:   "success",
		Date:     currentDate,
		Analysis: response,
	})
}

func callGemini2API(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.9, // Max creativity for realistic market simulation
			"topK":            40,
			"topP":            0.95,
			"maxOutputTokens": 8192,
		},
		"safetySettings": []map[string]interface{}{
			{
				"category":  "HARM_CATEGORY_DANGEROUS_CONTENT",
				"threshold": "BLOCK_NONE",
			},
			{
				"category":  "HARM_CATEGORY_HARASSMENT",
				"threshold": "BLOCK_NONE",
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request body: %v", err)
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-goog-api-key", apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Gemini 2.0 API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return "", fmt.Errorf("error from Gemini 2.0 API (status %d): %v", resp.StatusCode, errorResponse)
	}

	var geminiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini 2.0 response: %v", err)
	}

	if candidates, ok := geminiResp["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			if content, ok := candidate["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						if text, ok := part["text"].(string); ok {
							return text, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no content received from Gemini 2.0")
}
