package main

import (
	"context"
	"html/template"
	"king/app/service"
	"log"
	"net/http"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/api/option"
)

func mainHandler(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	return tmpl.Execute(c.Response().Writer, nil)
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}, error=${error}, bytes_in=${bytes_in}, bytes_out=${bytes_out}\n",
	}))

	// Initialize services
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	giftService := &service.GiftService{
		GenAIClient:        client,
		GoogleSearchAPIKey: os.Getenv("SEARCH_API_KEY"),
		GoogleSearchCX:     os.Getenv("SEARCH_CX"),
	}

	e.Static("/static", "static")
	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/home")
	})
	e.GET("/home", mainHandler)

	// Pass service to handler (using a closure or method)
	e.POST("/suggest-gift", func(c echo.Context) error {
		service.SuggestGiftHandler(c.Response().Writer, c.Request(), giftService)
		return nil
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	e.Logger.Infof("Starting server on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
