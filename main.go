// Go File Upload Server
//
// This server provides a web interface for uploading files with the following features:
// - Static HTML upload page with JavaScript-powered progress bar
// - Supports any file type with no restrictions
// - Handles large file uploads with streaming
// - Saves files to ./uploaded/ directory
// - Lists and allows downloading of uploaded files
// - Serves static assets from ./static/ directory
//
// Usage:
//   go run main.go
//
// The server will start on http://localhost:8080
// Access the upload interface at http://localhost:8080/
//
// Endpoints:
//   GET  /           - Upload form page
//   GET  /static/*   - Static assets (JS, CSS)
//   POST /upload     - File upload endpoint
//   GET  /files      - List uploaded files (JSON)
//   GET  /download/* - Download uploaded files

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	uploadDir  = "./uploaded"
	staticDir  = "./static"
	serverPort = ":8080"
)

// FileInfo represents information about an uploaded file
type FileInfo struct {
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"modTime"`
	DownloadURL string    `json:"downloadUrl"`
}

// HTML template for the upload page
const uploadPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Upload Server</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>File Upload Server</h1>
        
        <div class="upload-section">
            <h2>Upload File</h2>
            <form id="uploadForm" enctype="multipart/form-data">
                <div class="file-input-container">
                    <input type="file" id="fileInput" name="file" required>
                    <label for="fileInput" class="file-input-label">Choose File</label>
                    <span id="fileName" class="file-name"></span>
                </div>
                
                <div class="progress-container" id="progressContainer" style="display: none;">
                    <div class="progress-bar">
                        <div class="progress-fill" id="progressFill"></div>
                    </div>
                    <span class="progress-text" id="progressText">0%</span>
                </div>
                
                <button type="submit" id="uploadBtn">Upload File</button>
            </form>
            
            <div id="uploadStatus" class="upload-status"></div>
        </div>
        
        <div class="files-section">
            <h2>Uploaded Files</h2>
            <div id="filesList" class="files-list">
                <p class="loading">Loading files...</p>
            </div>
        </div>
    </div>
    
    <script src="/static/upload.js"></script>
</body>
</html>`

func main() {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Create static directory if it doesn't exist (for development)
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		log.Fatalf("Failed to create static directory: %v", err)
	}

	// Route handlers
	http.HandleFunc("/", handleUploadPage)
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/files", handleListFiles)
	http.HandleFunc("/download/", handleFileDownload)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	fmt.Printf("Starting file upload server on http://localhost%s\n", serverPort)
	fmt.Printf("Upload directory: %s\n", uploadDir)
	fmt.Printf("Static directory: %s\n", staticDir)

	log.Fatal(http.ListenAndServe(serverPort, nil))
}

// handleUploadPage serves the main upload page
func handleUploadPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("upload").Parse(uploadPageHTML)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// handleFileUpload handles the file upload with streaming support
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with no size limit (0 means no limit)
	if err := r.ParseMultipartForm(0); err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		log.Printf("Parse multipart form error: %v", err)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		log.Printf("Get file error: %v", err)
		return
	}
	defer file.Close()

	// Clean the filename to prevent directory traversal
	filename := filepath.Base(handler.Filename)
	if filename == "" || filename == "." || filename == ".." {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Create the destination file
	destPath := filepath.Join(uploadDir, filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create destination file", http.StatusInternalServerError)
		log.Printf("Create file error: %v", err)
		return
	}
	defer destFile.Close()

	// Stream the file content to destination
	_, err = io.Copy(destFile, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		log.Printf("Copy file error: %v", err)
		// Clean up the partially written file
		os.Remove(destPath)
		return
	}

	log.Printf("File uploaded successfully: %s (%d bytes)", filename, handler.Size)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":  true,
		"filename": filename,
		"size":     handler.Size,
		"message":  "File uploaded successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// handleListFiles returns a JSON list of all uploaded files
func handleListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := os.ReadDir(uploadDir)
	if err != nil {
		http.Error(w, "Failed to read upload directory", http.StatusInternalServerError)
		log.Printf("Read directory error: %v", err)
		return
	}

	var fileInfos []FileInfo
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				log.Printf("Get file info error for %s: %v", file.Name(), err)
				continue
			}

			fileInfos = append(fileInfos, FileInfo{
				Name:        file.Name(),
				Size:        info.Size(),
				ModTime:     info.ModTime(),
				DownloadURL: "/download/" + file.Name(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfos)
}

// handleFileDownload serves uploaded files for download
func handleFileDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filename from URL path
	filename := strings.TrimPrefix(r.URL.Path, "/download/")
	filename = filepath.Base(filename) // Prevent directory traversal

	if filename == "" || filename == "." || filename == ".." {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)
}
