package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Check current directory
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}
	fmt.Printf("Current directory: %s\n", pwd)

	// Check if go.mod exists
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("Error: go.mod file not found!")
		fmt.Println("Run 'go mod init cyberease' to initialize the module")
		return
	}

	// Read go.mod content
	content, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Printf("Error reading go.mod: %v\n", err)
		return
	}
	fmt.Println("\ngo.mod content:")
	fmt.Println(string(content))

	// Check project structure
	dirs := []string{"cmd", "scanner"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("Warning: %s directory not found!\n", dir)
			continue
		}

		// Check for main.go in cmd
		if dir == "cmd" {
			if _, err := os.Stat(filepath.Join(dir, "main.go")); os.IsNotExist(err) {
				fmt.Println("Warning: main.go not found in cmd directory!")
			}
		}

		// Check for scanner.go in scanner
		if dir == "scanner" {
			if _, err := os.Stat(filepath.Join(dir, "scanner.go")); os.IsNotExist(err) {
				fmt.Println("Warning: scanner.go not found in scanner directory!")
			}
		}
	}
}
