package stock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Structs
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type StockRecommendationResponse struct {
	Status   string `json:"status"`
	Date     string `json:"date"`
	Analysis string `json:"analysis"`
	Error    string `json:"error,omitempty"`
}

func DailyRecommendations(w http.ResponseWriter, r *http.Request) {
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

	currentDate := time.Now().Format("2006-01-02")

	prompt := fmt.Sprintf(`Anda adalah seorang analis saham Indonesia yang berpengalaman. Berikan rekomendasi saham harian untuk trading dengan brief konsisten berikut:

PROFIL TRADER:
- Modal: Rp 7,5 juta
- Target profit: 4-5%% per trade
- Frekuensi: 3x seminggu atau daily trading
- Strategi: Entry saat reversal, take profit cepat

TUGAS ANDA:
Berikan analisis mendalam untuk tanggal %s dan rekomendasikan 2-3 saham terbaik untuk entry hari ini dengan potensi profit 4-5%% besok/lusa.

FORMAT ANALISIS:
1. **RINGKASAN PASAR** - Kondisi IHSG, sentimen global, berita makro hari ini
2. **REKOMENDASI SAHAM** - Untuk setiap saham berikan:
   - Kode saham & nama perusahaan
   - Harga saat ini & target price
   - Alasan fundamental (laporan keuangan, berita, proyeksi)
   - Analisis teknikal (candlestick, support/resistance, volume, RSI, MACD)
   - Level entry yang tepat
   - Take profit & stop loss
   - Confidence level (High/Medium/Low)
3. **RISK MANAGEMENT** - Saran alokasi modal dan manajemen risiko
4. **TIMING** - Kapan waktu terbaik entry (opening, mid-day, closing)

KRITERIA SAHAM:
- Likuiditas tinggi (easy exit)
- Volatilitas cukup untuk profit 4-5%%
- Fundamental tidak bermasalah
- Pattern teknikal mendukung
- Volume trading memadai

Berikan analisis yang detail, praktis, dan actionable. Fokus pada saham-saham yang realistis bisa memberikan return 4-5%% dalam 1-2 hari trading.`, currentDate)

	response, err := callGeminiAPI(prompt)
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

func callGeminiAPI(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}

	geminiReqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(geminiReqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request body: %v", err)
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + apiKey
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Gemini API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return "", fmt.Errorf("error from Gemini API (status %d): %v", resp.StatusCode, errorResponse)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %v", err)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no content received from Gemini")
}
