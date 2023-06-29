package webserver

import (
	"net/http"

	"fmt"
	"log"
)

func Setup(port int, logger *log.Logger, errLogger *log.Logger) {
	http.HandleFunc("/api", simpleHandler)

	logger.Printf("Web server started on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		errLogger.Println("Failed to start the webserver on designated port:", err)
		return
	}
}

func simpleHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from Web Server!")
}
