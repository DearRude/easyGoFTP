package webserver

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type DiskUsageResponse struct {
	Directory string `json:"directory"`
	DiskUsage int64  `json:"diskUsage"`
	Error     string `json:"error,omitempty"`
}

func diskUsageHandler(c echo.Context) error {
	directory := c.Param("path")
	if directory == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Directory parameter is required")
	}

	responseChan := make(chan DiskUsageResponse)
	go func() {
		usage, err := calculateDiskUsage(directory)
		response := DiskUsageResponse{
			Directory: directory,
			DiskUsage: usage,
			Error:     "",
		}
		if err != nil {
			response.Error = err.Error()
		}
		responseChan <- response
	}()

	return c.JSON(http.StatusOK, <-responseChan)
}

func calculateDiskUsage(directory string) (int64, error) {
	var total int64

	directory = filepath.Join(cwd, directory)
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}
