package scanner

import (
	"encoding/json"
	"fmt"
	"html/template"
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
	ts, err := template.ParseFiles("themes/scanner.html")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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
