package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// --- STRUCTS UNTUK GEMINI API ---

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

type UserPromptRequest struct {
	Prompt string `json:"prompt"`
}

// --- STRUCTS UNTUK STOCK ANALYSIS ---

type StockAnalysisRequest struct {
	StockCode string `json:"stock_code"` // Untuk endpoint analisis saham spesifik
}

type StockRecommendationResponse struct {
	Status   string `json:"status"`
	Date     string `json:"date"`
	Analysis string `json:"analysis"`
	Error    string `json:"error,omitempty"`
}

// ------------------------------------

func main() {
	app := fiber.New(fiber.Config{
		// Untuk Vercel deployment
		DisableStartupMessage: true,
	})

	// CORS middleware untuk frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	setupRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}

// --- API Handlers ---

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Stock Analysis API with AI-powered recommendations",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"daily_recommendations": "GET /api/stock/daily-recommendations",
				"analyze_stock":         "POST /api/stock/analyze",
				"general_ai":            "POST /api/prompt",
			},
		})
	})

	// Health check untuk Vercel
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	api := app.Group("/api")

	// Routes Gemini (untuk general purpose)
	api.Post("/prompt", handleGeminiPrompt)

	// Routes untuk Stock Analysis
	api.Get("/stock/daily-recommendations", getDailyStockRecommendations)
	api.Post("/stock/analyze", analyzeSpecificStock)
}

func handleGeminiPrompt(c *fiber.Ctx) error {
	userReq := new(UserPromptRequest)
	if err := c.BodyParser(userReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if userReq.Prompt == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Prompt cannot be empty"})
	}

	response, err := callGeminiAPI(userReq.Prompt)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":   "success",
		"response": response,
	})
}

// --- STOCK ANALYSIS HANDLERS ---

// Endpoint 1: Rekomendasi saham harian dengan analisis mendalam
func getDailyStockRecommendations(c *fiber.Ctx) error {
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
		return c.Status(500).JSON(StockRecommendationResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(StockRecommendationResponse{
		Status:   "success",
		Date:     currentDate,
		Analysis: response,
	})
}

// Endpoint 2: Analisis saham spesifik
func analyzeSpecificStock(c *fiber.Ctx) error {
	req := new(StockAnalysisRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if req.StockCode == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Stock code cannot be empty"})
	}

	currentDate := time.Now().Format("2006-01-02")

	prompt := fmt.Sprintf(`Anda adalah seorang analis saham Indonesia yang berpengalaman. Analisis mendalam saham %s untuk trading dengan brief konsisten berikut:

PROFIL TRADER:
- Modal: Rp 7,5 juta
- Target profit: 4-5%% per trade
- Frekuensi: 3x seminggu atau daily trading
- Strategi: Entry saat reversal, take profit cepat

TUGAS ANDA:
Berikan analisis lengkap saham %s per tanggal %s. Apakah saham ini layak untuk entry hari ini/besok dengan target profit 4-5%%?

FORMAT ANALISIS:
1. **OVERVIEW SAHAM**
   - Nama perusahaan & sektor
   - Harga saat ini & pergerakan 1 minggu terakhir
   - Market cap & volume trading
   
2. **ANALISIS FUNDAMENTAL**
   - Kondisi laporan keuangan terbaru
   - Berita terkini yang mempengaruhi saham
   - Proyeksi bisnis & outlook industri
   - Faktor katalisa positif/negatif

3. **ANALISIS TEKNIKAL MENDALAM**
   - Trend jangka pendek (1-7 hari)
   - Support & resistance level
   - Candlestick pattern terbaru
   - Indikator teknikal: RSI, MACD, Volume
   - Fibonacci retracement (jika relevant)

4. **REKOMENDASI TRADING**
   - Apakah worth it untuk entry? (Ya/Tidak + reasoning)
   - Timing entry yang optimal
   - Entry price range
   - Take profit target (4-5%%)
   - Stop loss level
   - Confidence level (High/Medium/Low)
   - Alokasi modal yang disarankan

5. **RISK ASSESSMENT**
   - Risk level untuk saham ini
   - Faktor risiko yang perlu diwaspadai
   - Alternative action jika setup gagal

6. **KESIMPULAN**
   - Summary: BUY/HOLD/AVOID
   - Timeline holding (berapa hari)
   - Expected return realistis

Berikan analisis yang honest, detail, dan praktis. Jika saham tidak bagus untuk trading, katakan dengan jelas dan berikan alasannya.`, req.StockCode, req.StockCode, currentDate)

	response, err := callGeminiAPI(prompt)
	if err != nil {
		return c.Status(500).JSON(StockRecommendationResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(StockRecommendationResponse{
		Status:   "success",
		Date:     currentDate,
		Analysis: response,
	})
}

// Helper function untuk memanggil Gemini API
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
