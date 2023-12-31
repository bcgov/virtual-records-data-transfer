package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateFiles(t *testing.T) {
	// Set up a temporary source directory with test files
	sourceDir, err := ioutil.TempDir("", "source")
	tempDir := os.TempDir()
	fmt.Print(tempDir)
	if err != nil {
		t.Fatal("Error creating temporary source directory:", err)
	}
	defer os.RemoveAll(sourceDir)

	// Create some test files in the source directory
	for i := 1; i <= 3; i++ {
		filePath := filepath.Join(sourceDir, fmt.Sprintf("file%d.txt", i))
		err := ioutil.WriteFile(filePath, []byte(fmt.Sprintf("Content of virtual court file %d", i)), 0644)
		if err != nil {
			t.Fatalf("Error creating test file: %v", err)
		}
	}

	// Set up a temporary destination directory
	destinationDir, err := ioutil.TempDir("", "destination")
	if err != nil {
		t.Fatal("Error creating temporary destination directory:", err)
	}
	defer os.RemoveAll(destinationDir)

	// Run the migration
	err = migrateFiles(sourceDir, destinationDir, tempDir)
	if err != nil {
		t.Fatalf("Error during migration: %v", err)
	}

	fileCount, err := countFiles(destinationDir)
	if err != nil {
		t.Fatalf("Error reading destination directory: %v", err)
	}

	expectedFileCount := 3 // Adjust based on the number of test files created
	if fileCount != expectedFileCount {
		t.Errorf("Expected %d files in destination, got %d", expectedFileCount, fileCount)
	}

	// Test: Check the content of one of the migrated files
	firstDestFileContent, err := readContent(destinationDir, "file1.txt")
	if err != nil {
		t.Fatalf("Error reading content of destination file: %v", err)
	}

	expectedContent := []byte("Content of virtual court file 1")
	if string(firstDestFileContent) != string(expectedContent) {
		t.Errorf("Expected content '%s' in destination file, got '%s'", expectedContent, firstDestFileContent)
	}
}

func countFiles(dir string) (int, error) {
	var count int

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			count++
		}

		return nil
	})

	return count, err
}
func readContent(dir string, filename string) ([]byte, error) {
	var content []byte

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == filename {
			content1, err := os.ReadFile(path)
			content = content1

			if err != nil {
				return err
			}
		}

		return nil
	})

	return content, err
}
