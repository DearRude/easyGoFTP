package webserver

import (
	"net/http"

	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	logger    *log.Logger
	errLogger *log.Logger
	cwd       string
	db        *sql.DB
)

func Setup(port int, stdLogger *log.Logger, errorLogger *log.Logger, curr_dir string, database *sql.DB) {
	logger = stdLogger
	errLogger = errorLogger
	cwd = curr_dir
	db = database

	e := echo.New()
	api := e.Group("/api")

	api.Use(middleware.BasicAuth(loginAdmin))

	api.GET("/ping", pingHandler)

	api.GET("/disk-usage/:path", diskUsageHandler)
	api.GET("/directory-content/:path", directoryContentHandler)
	api.GET("/file-info/:path", fileInfoHandler)

	api.POST("/users", createUser)
	api.GET("/users/:id", getUser)
	api.PUT("/users/:id", updateUser)
	api.DELETE("/users/:id", deleteUser)

	logger.Printf("Web server started on port %d\n", port)
	if err := e.Start(fmt.Sprintf(":%d", port)); err != nil {
		errLogger.Fatalf("Failed to start the webserver on designated port: %v", err)
	}
}

type PingResponse struct {
	Message     string        `json:"message"`
	ElapsedTime time.Duration `json:"elapsedTime"`
}

func pingHandler(c echo.Context) error {
	start := time.Now()

	resp := map[string]interface{}{
		"message": "Pong!",
		"delay":   time.Since(start).String(),
	}
	return c.JSON(http.StatusOK, resp)
}
