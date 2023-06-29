package webserver

import (
	"net/http"
	"os"

	"path/filepath"
)

type DirectoryContentResponse struct {
	Directory string     `json:"directory"`
	FileInfos []FileInfo `json:"fileInfos"`
	Error     string     `json:"error,omitempty"`
}

func directoryContentHandler(w http.ResponseWriter, r *http.Request) {
	directory := r.URL.Query().Get("directory")
	if directory == "" {
		http.Error(w, "Directory parameter is required", http.StatusBadRequest)
		return
	}

	responseChan := make(chan DirectoryContentResponse)
	go func() {
		fileInfos, err := getDirectoryContent(directory)
		response := DirectoryContentResponse{
			Directory: directory,
			FileInfos: fileInfos,
			Error:     "",
		}
		if err != nil {
			response.Error = err.Error()
		}
		responseChan <- response
	}()

	sendJSONResponse(w, <-responseChan)
}

func getDirectoryContent(directory string) ([]FileInfo, error) {
	var fileInfos []FileInfo

	directory = filepath.Join(cwd, directory)
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}

		fileInfos = append(fileInfos, FileInfo{
			Name:    file.Name(),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().String(),
			IsDir:   info.IsDir(),
		})
	}

	return fileInfos, nil
}
