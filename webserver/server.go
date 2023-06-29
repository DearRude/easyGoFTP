package webserver

import (
	"net/http"

	"encoding/json"
	"fmt"
	"log"
	"time"
)

var (
	logger    *log.Logger
	errLogger *log.Logger
	cwd       string
)

func Setup(port int, stdLogger *log.Logger, errorLogger *log.Logger, curr_dir string) {
	logger = stdLogger
	errLogger = errorLogger
	cwd = curr_dir

	http.HandleFunc("/api/ping", pingHandler)

	http.HandleFunc("/api/disk-usage", diskUsageHandler)
	http.HandleFunc("/api/directory-content", directoryContentHandler)
	http.HandleFunc("/api/file-info", fileInfoHandler)

	logger.Printf("Web server started on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		errLogger.Println("Failed to start the webserver on designated port:", err)
		return
	}
}

type PingResponse struct {
	Message     string        `json:"message"`
	ElapsedTime time.Duration `json:"elapsedTime"`
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	response := PingResponse{
		Message:     "pong",
		ElapsedTime: time.Since(startTime),
	}

	sendJSONResponse(w, response)
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode JSON response: %s", err), http.StatusInternalServerError)
	}
}
