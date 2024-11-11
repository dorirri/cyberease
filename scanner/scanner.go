package scanner

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ScanResult struct {
	Filename string    `json:"filename"`
	IsClean  bool      `json:"is_clean"`
	Threat   string    `json:"threat,omitempty"`
	ScanTime time.Time `json:"scan_time"`
	Debug    string    `json:"debug,omitempty"`
}

type WindowsDefenderScanner struct {
	ScanPath string
}

func NewScanner() *WindowsDefenderScanner {
	return &WindowsDefenderScanner{
		ScanPath: os.TempDir(),
	}
}

func (s *WindowsDefenderScanner) ScanFile(filepath string) (*ScanResult, error) {
	result := &ScanResult{
		Filename: filepath,
		IsClean:  true,
		ScanTime: time.Now(),
	}

	log.Printf("Starting scan of file: %s", filepath)

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filepath)
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Start-MpScan -ScanPath '%s' -ScanType CustomScan", filepath))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Scan command error: %v - Output: %s", err, string(output))
		return nil, fmt.Errorf("scan failed: %v", err)
	}

	log.Printf("Scan completed. Output: %s", string(output))

	histCmd := exec.Command("powershell", "-Command",
		"Get-MpThreatDetection | Select-Object -Last 1")

	histOutput, err := histCmd.CombinedOutput()
	if err != nil {
		log.Printf("History check error: %v - Output: %s", err, string(histOutput))
	}

	if strings.Contains(string(histOutput), filepath) {
		result.IsClean = false
		result.Threat = "Potential threat detected"
	}

	result.Debug = fmt.Sprintf("Scan output: %s\nHistory output: %s", string(output), string(histOutput))
	return result, nil
}

func handleGetRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, UploadForm)
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing POST request from %s", r.RemoteAddr)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s", header.Filename)

	tempFile, err := os.CreateTemp("", "scan-*"+filepath.Ext(header.Filename))
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		log.Printf("Error copying file: %v", err)
		http.Error(w, "Error processing file", http.StatusInternalServerError)
		return
	}

	log.Printf("File saved to temp location: %s", tempFile.Name())

	scanner := NewScanner()
	result, err := scanner.ScanFile(tempFile.Name())
	if err != nil {
		log.Printf("Scan error: %v", err)
		http.Error(w, fmt.Sprintf("Scan failed: %v", err), http.StatusInternalServerError)
		return
	}

	result.Filename = header.Filename

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Scan completed successfully for file: %s", header.Filename)
}

func ScanHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetRequest(w)
	case http.MethodPost:
		handlePostRequest(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

const UploadForm = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Malware Scanner</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: 'Arial', sans-serif;
            background-color: #111;
            color: #fff;
        }

        header {
            background-color: #1e1e1e;
            padding: 20px;
            text-align: center;
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.5);
        }

        header h1 {
            color: #ffcc00;
            font-size: 36px;
            text-transform: uppercase;
            letter-spacing: 2px;
        }

        .upload-form {
            border: 2px dashed #ccc;
            padding: 20px;
            text-align: center;
            background-color: #1a1a1a;
            border-radius: 10px;
            margin: 20px auto;
            max-width: 600px;
            box-shadow: 0px 4px 15px rgba(0, 0, 0, 0.7);
        }

        .result {
            margin-top: 20px;
            padding: 15px;
            border-radius: 8px;
            white-space: pre-wrap;
        }

        .clean {
            background-color: #2b2b2b;
            border: 1px solid #b3ffb3;
        }

        .infected {
            background-color: #3b3b3b;
            border: 1px solid #ffb3b3;
        }

        .error {
            background-color: #333;
            border: 1px solid #ffeeba;
            color: #ffeeba;
        }

        button {
            background-color: #444;
            color: #fff;
            padding: 10px 20px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            margin-top: 10px;
            transition: background-color 0.3s ease;
        }

        button:hover {
            background-color: #ffcc00;
        }

        .loading {
            display: none;
            margin-top: 20px;
            color: #fff;
        }

        #debugInfo {
            margin-top: 20px;
            padding: 10px;
            background-color: #222;
            border: 1px solid #ddd;
            font-family: monospace;
            color: #fff;
        }

        /* Adjusting link styles */
        a {
            color: #ffcc00;
            text-decoration: none;
        }

        a:hover {
            color: #e6b800;
        }
    </style>
</head>
<body>
    <header>
        <h1>File Security Scanner</h1>
    </header>
    <p style="text-align: center;">Upload a file to scan it for potential security threats.</p>
    <div class="upload-form">
        <form id="scanForm" enctype="multipart/form-data">
            <input type="file" name="file" id="fileInput" required>
            <br>
            <button type="submit">Scan File</button>
        </form>
    </div>
    <div id="loading" class="loading">
        Scanning file... Please wait...
    </div>
    <div id="result" class="result" style="display: none;"></div>
    <div id="debugInfo">
        <strong>Debug Info:</strong>
        <pre id="debugOutput"></pre>
    </div>

    <script>
        function appendDebug(message) {
            const debugOutput = document.getElementById('debugOutput');
            const timestamp = new Date().toISOString();
            debugOutput.textContent += timestamp + ": " + message + "\n";
            console.log(message);
        }

        document.getElementById('fileInput').addEventListener('change', function(e) {
            const file = e.target.files[0];
            if (file) {
                appendDebug("File selected: " + file.name + " (" + file.size + " bytes, type: " + file.type + ")");
            }
        });

        document.getElementById('scanForm').onsubmit = async function(e) {
            e.preventDefault();
            appendDebug('Form submission started');
            
            const formData = new FormData(e.target);
            const file = formData.get('file');
            if (!file) {
                appendDebug('Error: No file selected');
                return;
            }
            appendDebug('Preparing to upload file: ' + file.name);
            
            const resultDiv = document.getElementById('result');
            const loadingDiv = document.getElementById('loading');
            const submitButton = e.target.querySelector('button');
            
            submitButton.disabled = true;
            loadingDiv.style.display = 'block';
            resultDiv.style.display = 'none';
            
            try {
                appendDebug('Starting file upload to server...');
                const response = await fetch('/scan', {
                    method: 'POST',
                    body: formData
                });
                
                appendDebug('Server responded with status: ' + response.status);
                
                if (!response.ok) {
                    const errorText = await response.text();
                    appendDebug('Server error: ' + errorText);
                    throw new Error(errorText || 'Server error');
                }
                
                appendDebug('Parsing server response...');
                const result = await response.json();
                appendDebug('Server response parsed successfully');
                
                resultDiv.className = 'result ' + (result.is_clean ? 'clean' : 'infected');
                resultDiv.innerHTML = 
                    '<h3>Scan Results:</h3>' +
                    '<p>File: ' + result.filename + '</p>' +
                    '<p>Status: ' + (result.is_clean ? 'Clean ✅' : 'Threat Detected ⚠️') + '</p>' +
                    (result.threat ? '<p>Threat: ' + result.threat + '</p>' : '') +
                    '<p>Scan Time: ' + new Date(result.scan_time).toLocaleString() + '</p>' +
                    (result.debug ? '<pre>Debug Info:\n' + result.debug + '</pre>' : '');
                
                appendDebug('Results displayed successfully');
                
            } catch (error) {
                appendDebug('Error occurred: ' + error.message);
                resultDiv.className = 'result error';
                resultDiv.innerHTML = 'Error scanning file: ' + error.message;
                console.error('Error:', error);
            } finally {
                submitButton.disabled = false;
                loadingDiv.style.display = 'none';
                resultDiv.style.display = 'block';
            }
        };

        // Log when page loads
        appendDebug('Scanner page loaded successfully');
    </script>
</body>
</html>
`
