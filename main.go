package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Product model (tetap sama)
type Product struct {
	gorm.Model
	Code  string `json:"code"`
	Price uint   `json:"price"`
}

// Database instance (tetap sama)
var DB *gorm.DB

// --- STRUCTS BARU UNTUK GEMINI API ---

// Struct untuk request yang dikirim ke Gemini
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

// Struct untuk response yang diterima dari Gemini
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

// Struct untuk request dari user ke API kita
type UserPromptRequest struct {
	Prompt string `json:"prompt"`
}

// ------------------------------------


func connectDatabase() {
	// ... (fungsi ini tetap sama, tidak perlu diubah)
	var err error
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, sslmode)
	
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
	}
	fmt.Println("Database connection successfully opened")
	DB.AutoMigrate(&Product{})
	fmt.Println("Database Migrated")
}

func main() {
	// ... (fungsi ini tetap sama, tidak perlu diubah)
	connectDatabase()
	app := fiber.New()
	setupRoutes(app)
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}

// --- API Handlers ---

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Selamat datang di API Go Fiber dengan Postgres!"})
	})
	
	api := app.Group("/api")

	// Routes Produk (tetap sama)
	api.Get("/products", getProducts)
	api.Post("/products", createProduct)

	// --- ROUTE BARU UNTUK GEMINI ---
	api.Post("/prompt", handleGeminiPrompt)
}

// ... (fungsi getProducts dan createProduct tetap sama)

func getProducts(c *fiber.Ctx) error {
	var products []Product
	if err := DB.Find(&products).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not get products", "data": err})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Products found", "data": products})
}

func createProduct(c *fiber.Ctx) error {
	product := new(Product)
	if err := c.BodyParser(product); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	if err := DB.Create(&product).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create product", "data": err})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "Product created", "data": product})
}

// --- HANDLER BARU UNTUK GEMINI ---

func handleGeminiPrompt(c *fiber.Ctx) error {
	// 1. Ambil prompt dari request body user
	userReq := new(UserPromptRequest)
	if err := c.BodyParser(userReq); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if userReq.Prompt == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Prompt cannot be empty"})
	}

	// 2. Ambil API Key dari environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(500).JSON(fiber.Map{"error": "GEMINI_API_KEY is not set"})
	}
	
	// 3. Siapkan request untuk dikirim ke Gemini API
	geminiReqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: userReq.Prompt},
				},
			},
		},
	}
	
	// Ubah struct menjadi JSON
	jsonBody, err := json.Marshal(geminiReqBody)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create request body"})
	}

	// 4. Buat dan kirim HTTP POST request ke Gemini
	// Menggunakan model gemini-1.5-flash yang lebih baru dan efisien
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + apiKey
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create HTTP request"})
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send request to Gemini API"})
	}
	defer resp.Body.Close()

	// 5. Baca dan parse response dari Gemini
	if resp.StatusCode != http.StatusOK {
		// Jika ada error dari Gemini, tampilkan pesannya
		var errorResponse map[string]interface{}
    	json.NewDecoder(resp.Body).Decode(&errorResponse)
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": "Error from Gemini API", "details": errorResponse})
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse Gemini response"})
	}
	
	// 6. Ekstrak jawaban teks dan kirim kembali ke user
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return c.JSON(fiber.Map{
			"status": "success",
			"response": geminiResp.Candidates[0].Content.Parts[0].Text,
		})
	}

	return c.Status(500).JSON(fiber.Map{"error": "No content received from Gemini"})
}