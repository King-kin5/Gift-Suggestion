package main

import (
	"net/http"
	"os"
	"html/template"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"king/app/service"
)

func mainHandler(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	
	return tmpl.Execute(c.Response().Writer,nil)
}


func main()  {
	e:=echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}, error=${error}, bytes_in=${bytes_in}, bytes_out=${bytes_out}\n",
	}))
	

	e.Static("/static","static")
	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/home")
	})
	e.GET("/home", mainHandler)
	e.POST("/suggest-gift", echo.WrapHandler(http.HandlerFunc(service.SuggestGiftHandler)))
		// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	e.Logger.Infof("Starting server on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}