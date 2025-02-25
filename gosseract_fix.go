// gosseract_fix.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Path to the gosseract module
	goModPath := "/Users/thilobarth/go/pkg/mod/github.com/otiai10/gosseract/v2@v2.4.1"
	
	// Path to the tessbridge.cpp file
	tessbridgePath := filepath.Join(goModPath, "tessbridge.cpp")
	
	// Read the file
	content, err := os.ReadFile(tessbridgePath)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}
	
	// Create a backup
	backupPath := tessbridgePath + ".bak"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		fmt.Printf("Error creating backup: %s\n", err)
		return
	}
	
	// Modify the file to use the correct paths for Homebrew on Mac
	modified := strings.Replace(string(content),
		"#include <leptonica/allheaders.h>",
		"#include \"/opt/homebrew/Cellar/leptonica/1.85.0/include/leptonica/allheaders.h\"",
		1)
	
	modified = strings.Replace(modified,
		"#include <tesseract/baseapi.h>",
		"#include \"/opt/homebrew/Cellar/tesseract/5.5.0/include/tesseract/baseapi.h\"",
		1)
	
	// Write the modified content back to the file
	if err := os.WriteFile(tessbridgePath, []byte(modified), 0644); err != nil {
		fmt.Printf("Error writing modified file: %s\n", err)
		return
	}
	
	fmt.Println("Successfully patched tessbridge.cpp")
}
