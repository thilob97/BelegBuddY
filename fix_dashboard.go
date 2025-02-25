package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Path to dashboard.go
	filePath := "ui/views/dashboard.go"
	
	// Read the file
	file, err := os.Open(filePath)
	if err \!= nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read line by line
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Fix for invalid nesting of Style in XAxis
	inXAxis := false
	removeNextLine := false
	var fixedLines []string
	
	for i, line := range lines {
		if removeNextLine {
			removeNextLine = false
			continue
		}
		
		trimmedLine := strings.TrimSpace(line)
		
		// Track if we're inside XAxis declaration
		if strings.Contains(trimmedLine, "XAxis: chart.Style{") {
			inXAxis = true
			fixedLines = append(fixedLines, line)
			continue
		}
		
		// Handle nested Style within XAxis
		if inXAxis && strings.Contains(trimmedLine, "Style: chart.Style{") {
			// Replace the line with direct property assignment
			indentation := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			fixedLines[len(fixedLines)-1] = indentation + "XAxis: chart.Style{"
			removeNextLine = true
			continue
		}
		
		// Exit XAxis block
		if inXAxis && trimmedLine == "}," {
			inXAxis = false
		}
		
		fixedLines = append(fixedLines, line)
	}

	// Write the fixed file
	outFile, err := os.Create(filePath)
	if err \!= nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, line := range fixedLines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()
	
	fmt.Println("Fixed the dashboard.go file")
}
