// static/upload.js
document.addEventListener("DOMContentLoaded", function() {
    const form = document.getElementById("uploadForm");
    const fileInput = document.getElementById("fileInput");
    const progressContainer = document.getElementById("progressContainer");
    const progressBar = document.getElementById("progressBar");
    const progressPercent = document.getElementById("progressPercent");
    const result = document.getElementById("result");

    form.addEventListener("submit", function(e) {
        e.preventDefault();
        const file = fileInput.files[0];
        if (!file) return;

        const xhr = new XMLHttpRequest();
        xhr.open("POST", "/upload", true);

        xhr.upload.onprogress = function(e) {
            if (e.lengthComputable) {
                progressContainer.style.display = "block";
                let percent = Math.round((e.loaded / e.total) * 100);
                progressBar.value = percent;
                progressPercent.textContent = percent + "%";
            }
        };

        xhr.onload = function() {
            progressContainer.style.display = "none";
            if (xhr.status === 200 && xhr.responseText === "success") {
                result.textContent = "Upload successful! Reloading file list...";
                setTimeout(() => window.location.reload(), 1000);
            } else {
                result.textContent = "Upload failed: " + xhr.responseText;
            }
        };

        xhr.onerror = function() {
            result.textContent = "Upload failed. Network error.";
        };

        const formData = new FormData();
        formData.append("file", file);
        xhr.send(formData);
    });
});
