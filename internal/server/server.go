package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"blog-platform/internal/database"
)

type Server struct {
	Port int
	DB   database.BlogRepository
}

func NewServer() *http.Server {
	dbSettings := database.Settings{
		HostName:   os.Getenv("DB_HOST"),
		Port:       os.Getenv("DB_PORT"),
		Username:   os.Getenv("DB_USERNAME"),
		Password:   os.Getenv("DB_PASSWORD"),
		DbName:     os.Getenv("DB_DATABASE"),
		AuthSource: os.Getenv("DB_AUTHSOURCE"),
	}
	db, err := database.New(dbSettings)

	if err != nil {
		log.Fatal(err)
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))

	if err != nil {
		log.Fatal(err)
	}

	NewServer := &Server{
		Port: port,
		DB:   db,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.Port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/health", s.HealthHandler)
	e.GET("/posts", s.GetBlogsHandler)
	e.GET("/posts/:id", s.GetBlogHandler)
	e.POST("/posts", s.CreateBlogHandler)
	e.PUT("/posts/:id", s.UpdateBlogHandler)
	e.DELETE("/posts/:id", s.DeleteBlogHandler)

	return e
}
