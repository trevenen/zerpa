# Zerpa - Go File Upload Server

A simple, lightweight Go web server for file uploads with a modern web interface.

## Features

- **Easy File Uploads**: Drag-and-drop or click to select files for upload
- **Progress Tracking**: Real-time upload progress bar with percentage display
- **No Restrictions**: Upload any file type with no size limits
- **Large File Support**: Streaming upload support for large files
- **File Management**: View and download all uploaded files
- **Modern UI**: Clean, responsive web interface
- **Static Assets**: Organized separation of CSS/JS from Go code

## Quick Start

1. **Clone and run**:
   ```bash
   git clone https://github.com/trevenen/zerpa.git
   cd zerpa
   go run main.go
   ```

2. **Access the application**:
   Open your browser and navigate to `http://localhost:8080`

3. **Upload files**:
   - Select files using the "Choose File" button
   - Watch the progress bar during upload
   - View and download uploaded files in the list below

## Project Structure

```
zerpa/
├── main.go              # Go web server with all endpoints
├── static/
│   ├── upload.js        # JavaScript for AJAX uploads and progress
│   └── style.css        # CSS styling for the web interface
├── uploaded/            # Directory for uploaded files (created automatically)
└── README.md           # This file
```

## API Endpoints

- `GET /` - Upload form page (HTML interface)
- `GET /static/*` - Static assets (CSS, JavaScript)
- `POST /upload` - File upload endpoint (accepts multipart/form-data)
- `GET /files` - List uploaded files (JSON response)
- `GET /download/{filename}` - Download specific uploaded file

## Configuration

The server runs on port 8080 by default. To change this, modify the `serverPort` constant in `main.go`:

```go
const serverPort = ":8080"  // Change to your preferred port
```

## File Storage

- Uploaded files are stored in the `./uploaded/` directory
- The directory is created automatically if it doesn't exist
- Files retain their original names (sanitized for security)
- No authentication or restrictions on file types

## Development

Built with:
- **Go** (standard library only, no external dependencies)
- **Vanilla JavaScript** (no frameworks)
- **CSS3** (responsive design)

## Security Notes

- File names are sanitized to prevent directory traversal
- Files are streamed to disk to handle large uploads efficiently
- No authentication is implemented (suitable for local/trusted environments)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
