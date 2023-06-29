package webserver

import (
	"net/http"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"modTime"`
	IsDir   bool   `json:"isDir"`
}

type FileInfoResponse struct {
	FileInfo FileInfo `json:"fileInfo"`
	Error    string   `json:"error,omitempty"`
}

func fileInfoHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "File parameter is required", http.StatusBadRequest)
		return
	}

	responseChan := make(chan FileInfoResponse)
	go func() {
		fileInfo, err := getFileInformation(filePath)
		response := FileInfoResponse{
			FileInfo: fileInfo,
			Error:    "",
		}
		if err != nil {
			response.Error = err.Error()
		}
		responseChan <- response
	}()

	sendJSONResponse(w, <-responseChan)
}

func getFileInformation(filePath string) (FileInfo, error) {
	filePath = filepath.Join(cwd, filePath)
	info, err := os.Stat(filePath)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Name:    filepath.Base(filePath),
		Size:    info.Size(),
		Mode:    info.Mode().String(),
		ModTime: info.ModTime().String(),
		IsDir:   info.IsDir(),
	}, nil
}
