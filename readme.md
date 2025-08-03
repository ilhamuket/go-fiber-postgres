# Stock Analysis API

API untuk analisis saham dengan AI-powered recommendations menggunakan Gemini AI.

## Features

- 🤖 **AI-Powered Analysis** - Menggunakan Google Gemini untuk analisis mendalam
- 📈 **Daily Recommendations** - Rekomendasi saham harian otomatis
- 🔍 **Specific Stock Analysis** - Analisis detail untuk saham tertentu
- 💰 **Trading Focused** - Disesuaikan untuk modal Rp 7,5 juta dengan target profit 4-5%
- 🚀 **Cloud Ready** - Siap deploy ke Vercel, Docker, atau cloud platform lainnya

## Quick Start

### Local Development

1. **Clone & Setup**
   ```bash
   git clone <your-repo>
   cd stock-analysis-api
   cp .env.example .env
   # Edit .env dengan GEMINI_API_KEY Anda
   ```

2. **Run with Go**
   ```bash
   go mod tidy
   go run main.go
   ```

3. **Or Run with Docker**
   ```bash
   docker-compose up --build
   ```

### Environment Variables

```bash
# Required
GEMINI_API_KEY=your_gemini_api_key_here

# Optional
PORT=3000
GO_ENV=development
```

## API Endpoints

### Stock Analysis
- `GET /api/stock/daily-recommendations` - Rekomendasi saham harian
- `POST /api/stock/analyze` - Analisis saham spesifik

### General
- `GET /` - API info
- `GET /api/health` - Health check
- `POST /api/prompt` - General AI chat

## Example Usage

### Daily Recommendations
```bash
curl https://your-api.vercel.app/api/stock/daily-recommendations
```

### Analyze Specific Stock
```bash
curl -X POST https://your-api.vercel.app/api/stock/analyze \
  -H "Content-Type: application/json" \
  -d '{"stock_code": "BBRI"}'
```

## Deployment

### Vercel (Recommended - Free)
```bash
npm i -g vercel
vercel --prod
```

### Docker
```bash
docker-compose up -d
```

### Manual
```bash
go build -o main
./main
```

## Trading Profile

API ini disesuaikan untuk profile trading:
- Modal: Rp 7,5 juta
- Target profit: 4-5% per trade  
- Frekuensi: 3x seminggu atau daily trading
- Strategi: Entry saat reversal, take profit cepat

## Tech Stack

- **Backend**: Go + Fiber
- **AI**: Google Gemini 1.5 Flash
- **Deploy**: Vercel / Docker
- **No Database** - Stateless untuk efisiensi

---

# .env.example
# Gemini AI Configuration
GEMINI_API_KEY=your_gemini_api_key_here

# App Configuration  
PORT=3000

# Optional
GO_ENV=development