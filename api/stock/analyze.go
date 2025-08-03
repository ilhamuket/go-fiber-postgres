package stock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type StockAnalysisRequest struct {
	StockCode string `json:"stock_code"`
}

type StockRecommendationResponse struct {
	Status   string `json:"status"`
	Date     string `json:"date"`
	Analysis string `json:"analysis"`
	Error    string `json:"error,omitempty"`
}

func Analyze(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, X-goog-api-key")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StockAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot parse JSON"})
		return
	}

	if req.StockCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Stock code cannot be empty"})
		return
	}

	currentDate := time.Now().Format("2006-01-02")
	stockContext := getStockContext(req.StockCode)

	prompt := fmt.Sprintf(`Anda adalah senior portfolio manager dari investment firm terkemuka di Jakarta dengan akses ke Bloomberg terminal dan data real-time. Klien Anda meminta analisis trading untuk saham %s pada %s.

%s

IMPORTANT CONTEXT: Ini adalah analisis internal untuk klien professional dengan modal Rp 7.5 juta. Berikan analisis praktis dan actionable.

TRADING PROFILE:
- Capital: Rp 7,500,000
- Target: 4-5%% profit per trade  
- Style: Active day/swing trading
- Risk tolerance: Medium-aggressive

ANALISIS PROFESIONAL UNTUK %s:

**STOCK DATA CURRENT**
- Company: %s
- Sector: [Based on your knowledge]
- Current price: Rp [Provide realistic estimate]
- Daily volume: [Typical volume for this stock]
- Market cap: [Calculate based on shares outstanding]

**TECHNICAL ANALYSIS**
- Trend: [Current short-term trend]
- Support levels: Rp [2 key levels]
- Resistance levels: Rp [2 key levels]  
- RSI (14): [Estimate current level]
- MACD status: [Above/below signal line]
- Volume pattern: [Recent volume vs average]

**FUNDAMENTAL SNAPSHOT**
- Recent earnings: [Latest quarter performance]
- Revenue growth: [YoY growth rate]
- Industry outlook: [Sector conditions]
- Key catalysts: [Upcoming events/news]

**TRADING RECOMMENDATION**

Entry Decision: [BUY/HOLD/AVOID]

If BUY:
- Entry zone: Rp [specific range]
- Target 1 (4%%): Rp [exact price]
- Target 2 (5%%): Rp [exact price] 
- Stop loss: Rp [price level]
- Position size: Rp [amount from 7.5M]
- Timeline: [1-3 days]

If AVOID:
- Reason: [Specific issues]
- Wait for: [Better conditions]
- Alternative: [Better stock picks]

**RISK FACTORS**
- Volatility: [High/Medium/Low]
- Liquidity: [Easy/Difficult to exit]
- Market correlation: [Beta estimate]

**EXECUTION PLAN**
- Best entry time: [Market hours preference]
- Order type: [Market/Limit recommendation]
- Monitoring: [Key levels to watch]

Confidence: [1-10] with rationale

Provide practical, actionable analysis based on current market knowledge for Indonesian stocks. Focus on realistic price levels and executable strategy for Rp 7.5M capital.`, req.StockCode, currentDate, stockContext, req.StockCode, getCompanyName(req.StockCode))

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

func getCompanyName(stockCode string) string {
	companies := map[string]string{
		"CDIA": "PT Chandra Daya Investasi Tbk",
		"GOTO": "PT GoTo Gojek Tokopedia Tbk",
		"BBCA": "PT Bank Central Asia Tbk",
		"BBRI": "PT Bank Rakyat Indonesia Tbk",
		"BMRI": "PT Bank Mandiri Tbk",
		"ASII": "PT Astra International Tbk",
		"UNVR": "PT Unilever Indonesia Tbk",
		"TLKM": "PT Telkom Indonesia Tbk",
		"COIN": "PT Digital Coin Indonesia Tbk",
		"CUAN": "PT Arha Capital Tbk",
	}

	if name, exists := companies[stockCode]; exists {
		return name
	}
	return fmt.Sprintf("PT %s Tbk", stockCode)
}

func getStockContext(stockCode string) string {
	switch stockCode {
	case "CDIA":
		return `CURRENT MARKET DATA (Chandra Daya Investasi):
Recent IPO with strong post-listing performance. Infrastructure/energy sector play, subsidiary of TPIA (Chandra Asri). Price range: 1,400-1,700 area based on recent trading. High volatility post-IPO typical. Multiple auto rejection atas (ARA) events. Strong fundamental backing from parent company. High retail interest. Trading volume varies significantly.`

	case "GOTO":
		return `CURRENT MARKET DATA (GoTo Gojek Tokopedia):
Established tech stock, large cap with high liquidity. Super app ecosystem business model. Typical trading range 100-150 based on historical patterns. Medium volatility suitable for swing trading. High daily volume, easy entry/exit. Focus on path to profitability, strong user metrics.`

	case "BBCA":
		return `CURRENT MARKET DATA (Bank Central Asia):
Premium Indonesian bank, highest quality banking stock. Typical range 8,000-10,000 based on historical. Low-medium volatility. Excellent liquidity. Consistent dividend payer. Strong digital banking. Defensive play with quality fundamentals.`

	case "BBRI":
		return `CURRENT MARKET DATA (Bank Rakyat Indonesia):
Large government-related bank, strong SME/rural network. Typical range 4,000-5,500. Low-medium volatility. High liquidity. Government backing provides stability. Solid dividend history.`

	case "CUAN":
		return `CURRENT MARKET DATA (Arha Capital):
Digital asset/crypto-related investment company. High volatility correlated with crypto markets. Speculative stock with high beta. Suitable for aggressive momentum traders.`

	case "COIN":
		return `CURRENT MARKET DATA (Digital Coin):
Crypto-related business model. Extreme volatility following crypto market sentiment. High risk, high reward potential. Momentum-driven trading.`

	default:
		return `MARKET DATA: Analyze based on sector characteristics and provide realistic price estimates for Indonesian market conditions.`
	}
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
			"temperature":     0.9, // Higher creativity for realistic estimates
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
			{
				"category":  "HARM_CATEGORY_HATE_SPEECH",
				"threshold": "BLOCK_NONE",
			},
			{
				"category":  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
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
