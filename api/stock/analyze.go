package stock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type StockAnalysisRequest struct {
	StockCode string `json:"stock_code"`
}

func Analyze(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
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
