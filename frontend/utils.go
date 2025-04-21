package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadAPIKey reads the file and returns the API key.
func ReadAPIKey(filename string) (string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return "", err // Handle file open errors
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for a line that starts with "API_KEY="
		if strings.HasPrefix(line, "API_KEY=") {
			// Extract the key after "API_KEY="
			return strings.TrimPrefix(line, "API_KEY="), nil
		}
	}

	// Handle case where no API_KEY was found
	if err := scanner.Err(); err != nil {
		return "", err // Handle scanner error
	}

	return "", fmt.Errorf("API key not found in the file")
}
