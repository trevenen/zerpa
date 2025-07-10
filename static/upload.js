// JavaScript for AJAX file upload with progress bar
// Handles file selection, upload progress, and file list management

document.addEventListener('DOMContentLoaded', function() {
    const uploadForm = document.getElementById('uploadForm');
    const fileInput = document.getElementById('fileInput');
    const fileName = document.getElementById('fileName');
    const progressContainer = document.getElementById('progressContainer');
    const progressFill = document.getElementById('progressFill');
    const progressText = document.getElementById('progressText');
    const uploadBtn = document.getElementById('uploadBtn');
    const uploadStatus = document.getElementById('uploadStatus');
    const filesList = document.getElementById('filesList');

    // Initialize the page
    loadFilesList();

    // Handle file input change
    fileInput.addEventListener('change', function() {
        if (this.files.length > 0) {
            const file = this.files[0];
            fileName.textContent = `${file.name} (${formatFileSize(file.size)})`;
            uploadStatus.textContent = '';
            uploadStatus.className = 'upload-status';
        } else {
            fileName.textContent = '';
        }
    });

    // Handle form submission
    uploadForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        const file = fileInput.files[0];
        if (!file) {
            showStatus('Please select a file first.', 'error');
            return;
        }

        uploadFile(file);
    });

    function uploadFile(file) {
        const formData = new FormData();
        formData.append('file', file);

        const xhr = new XMLHttpRequest();

        // Show progress container
        progressContainer.style.display = 'block';
        uploadBtn.disabled = true;
        uploadBtn.textContent = 'Uploading...';

        // Track upload progress
        xhr.upload.addEventListener('progress', function(e) {
            if (e.lengthComputable) {
                const percentComplete = (e.loaded / e.total) * 100;
                updateProgress(percentComplete);
            }
        });

        // Handle upload completion
        xhr.addEventListener('load', function() {
            if (xhr.status === 200) {
                try {
                    const response = JSON.parse(xhr.responseText);
                    if (response.success) {
                        showStatus(`File "${response.filename}" uploaded successfully!`, 'success');
                        resetForm();
                        loadFilesList(); // Refresh the files list
                    } else {
                        showStatus('Upload failed: ' + (response.message || 'Unknown error'), 'error');
                    }
                } catch (e) {
                    showStatus('Upload completed but response was invalid.', 'error');
                }
            } else {
                showStatus(`Upload failed: Server responded with status ${xhr.status}`, 'error');
            }
            
            resetUploadState();
        });

        // Handle upload errors
        xhr.addEventListener('error', function() {
            showStatus('Upload failed: Network error occurred.', 'error');
            resetUploadState();
        });

        // Handle upload abort
        xhr.addEventListener('abort', function() {
            showStatus('Upload was cancelled.', 'error');
            resetUploadState();
        });

        // Start the upload
        xhr.open('POST', '/upload');
        xhr.send(formData);
    }

    function updateProgress(percent) {
        const roundedPercent = Math.round(percent);
        progressFill.style.width = roundedPercent + '%';
        progressText.textContent = roundedPercent + '%';
    }

    function resetUploadState() {
        uploadBtn.disabled = false;
        uploadBtn.textContent = 'Upload File';
        progressContainer.style.display = 'none';
        updateProgress(0);
    }

    function resetForm() {
        uploadForm.reset();
        fileName.textContent = '';
        resetUploadState();
    }

    function showStatus(message, type) {
        uploadStatus.textContent = message;
        uploadStatus.className = `upload-status ${type}`;
    }

    function loadFilesList() {
        fetch('/files')
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then(files => {
                displayFilesList(files);
            })
            .catch(error => {
                console.error('Error loading files:', error);
                filesList.innerHTML = '<p class="error">Failed to load files list.</p>';
            });
    }

    function displayFilesList(files) {
        if (!files || files.length === 0) {
            filesList.innerHTML = '<p class="no-files">No files uploaded yet.</p>';
            return;
        }

        // Sort files by modification time (newest first)
        files.sort((a, b) => new Date(b.modTime) - new Date(a.modTime));

        const filesHTML = files.map(file => {
            const modTime = new Date(file.modTime).toLocaleString();
            return `
                <div class="file-item">
                    <div class="file-info">
                        <div class="file-name">${escapeHtml(file.name)}</div>
                        <div class="file-details">
                            <span class="file-size">${formatFileSize(file.size)}</span>
                            <span class="file-date">${modTime}</span>
                        </div>
                    </div>
                    <div class="file-actions">
                        <a href="${file.downloadUrl}" class="download-btn" download="${file.name}">Download</a>
                    </div>
                </div>
            `;
        }).join('');

        filesList.innerHTML = filesHTML;
    }

    function formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
});