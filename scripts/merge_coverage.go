package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	// Create combined coverage file
	combinedFile, err := os.Create("coverage_combined.out")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating combined coverage file: %v\n", err)
		os.Exit(1)
	}
	defer combinedFile.Close()

	// Write mode header
	fmt.Fprintln(combinedFile, "mode: atomic")

	// Merge coverage files if they exist
	files := []string{"coverage.out", "coverage_obsv.out"}

	for _, filename := range files {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			continue // Skip if file doesn't exist
		}

		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", filename, err)
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// Skip mode lines
			if len(line) > 5 && line[:5] == "mode:" {
				continue
			}
			fmt.Fprintln(combinedFile, line)
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
		}
	}

	// Close the combined file before copying
	combinedFile.Close()

	// Wait a moment for file handles to be released
	time.Sleep(100 * time.Millisecond)

	// Copy combined file to coverage.out (more robust than rename)
	if err := copyFile("coverage_combined.out", "coverage.out"); err != nil {
		fmt.Fprintf(os.Stderr, "Error copying combined coverage file: %v\n", err)
		os.Exit(1)
	}

	// Clean up temporary file
	os.Remove("coverage_combined.out")

	fmt.Println("Coverage files merged successfully")
}

// copyFile copies the contents of src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
