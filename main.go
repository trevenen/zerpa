// main.go
// Run with: go run main.go
// Visit http://localhost:8080 in your browser.

package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	uploadPath = "./uploaded"
	staticPath = "./static"
	maxMemory  = 1024 * 1024 * 512 // 512MB for parsing form (file is streamed to disk)
)

func main() {
	// Ensure upload directory exists
	err := os.MkdirAll(uploadPath, os.ModePerm)
	if err != nil {
		panic("Could not create upload directory: " + err.Error())
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	http.Handle("/uploaded/", http.StripPrefix("/uploaded/", http.FileServer(http.Dir(uploadPath))))
	http.HandleFunc("/", uploadPage)
	http.HandleFunc("/upload", uploadHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func uploadPage(w http.ResponseWriter, r *http.Request) {
	files, err := listUploadedFiles()
	if err != nil {
		http.Error(w, "Cannot list files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("upload").Parse(htmlPage))
	tmpl.Execute(w, files)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		http.Error(w, "Could not parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Could not get uploaded file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dstPath := filepath.Join(uploadPath, filepath.Base(handler.Filename))
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Could not save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Could not write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("success"))
}

func listUploadedFiles() ([]string, error) {
	f, err := os.Open(uploadPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fileInfos, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, fi := range fileInfos {
		if !fi.IsDir() {
			files = append(files, fi.Name())
		}
	}
	return files, nil
}

const htmlPage = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Go File Upload with Progress</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <h2>Upload a file</h2>
    <form id="uploadForm">
        <input type="file" name="file" id="fileInput" required>
        <button type="submit">Upload</button>
    </form>
    <div id="progressContainer" style="display:none;">
        <progress id="progressBar" value="0" max="100"></progress>
        <span id="progressPercent">0%</span>
    </div>
    <div id="result"></div>

    <h3>Uploaded files</h3>
    <ul>
    {{range .}}
        <li><a href="/uploaded/{{.}}" target="_blank">{{.}}</a></li>
    {{else}}
        <li>No files uploaded yet.</li>
    {{end}}
    </ul>
    <script src="/static/upload.js"></script>
</body>
</html>
`
