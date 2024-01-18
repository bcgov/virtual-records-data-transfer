package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var goroutineCounter int64

func main() {
	// Set up zerolog with pretty formatting for console output
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	base := os.Getenv("CIFS_PATH")
	sourcePath := base + os.Getenv("SOURCE_PATH")
	destinationPath := os.Getenv("DESTINATION_PATH")

	// Check if the source path is empty
	if os.Getenv("SOURCE_PATH") == "" || os.Getenv("CIFS_PATH") == "" {
		log.Fatal().Msg("Source path or CIFS path is empty")
	}

	//sourcePath := base + "Chunk_Jr\\TESTING"
	//destinationPath := "H:\\destination"

	log.Info().Msgf("Source path %v is mapped to smb %v", sourcePath, os.Getenv("SMB_SERVER")+os.Getenv("SMB_SHARE")+os.Getenv("SOURCE_PATH"))
	log.Info().Msgf("Destination path %v is mapped to s3 bucket %v on %v", destinationPath, os.Getenv("BUCKET_NAME"), os.Getenv("S3_ENDPOINT"))

	log.Info().Str("sourcePath", sourcePath).Str("destinationPath", destinationPath).Msg("Starting migration")

	err := migrateFiles(sourcePath, destinationPath, base)
	if err != nil {
		log.Fatal().Err(err).Msg("Error during migration")
	}

	// for {
	// 	log.Info().Msg(sourcePath)

	// 	time.Sleep(time.Second)
	// }
	//log.Info().Msg("Migration completed successfully.")
}

func migrateFiles(sourcePath, destinationPath string, prefix string) error {

	// Check if the destination directory exists
	if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
		return fmt.Errorf("destination directory does not exist: %s", destinationPath)
	}

	var wg sync.WaitGroup

	err := filepath.Walk(sourcePath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warn().Err(err).Str("filePath", filePath).Msg("Error accessing file")
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		relativePath := strings.TrimPrefix(filePath, prefix)

		if err != nil {
			log.Warn().Err(err).Str("filePath", filePath).Msg("Error creating relative path")
			return nil
		}
		destinationFilePath := filepath.Join(destinationPath, relativePath)

		err = os.MkdirAll(filepath.Dir(destinationFilePath), 0755)
		if err != nil {
			log.Fatal().Err(err).Str("filePath", filePath).Msg("Error creating destination directory")
			return nil
		}

		wg.Add(1)
		goroutineNum := atomic.AddInt64(&goroutineCounter, 1)

		go func(src, dest string, num int64) {
			defer wg.Done()
			log.Info().Int64("goroutineNum", num).Str("filePath", src).Msg("Processing file")
			err := copyFile(src, dest, num)
			if err != nil {
				log.Error().Err(err).Int64("goroutineNum", num).Str("filePath", src).Str("destinationPath", dest).Msg("Error copying file")
			} else {
				log.Info().Int64("goroutineNum", num).Str("filePath", src).Str("destinationPath", dest).Msg("File copied successfully")
			}
		}(filePath, destinationFilePath, goroutineNum)

		return nil
	})

	wg.Wait()

	if err != nil {
		log.Warn().Err(err).Msg("Migration completed with errors")
	} else {
		log.Info().Msg("Migration completed successfully.")
	}

	return err
}

func copyFile(source, destination string, num int64) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("error opening source file %s: %v", source, err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("error creating destination file %s: %v", destination, err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error in goroutine #%d copying from %s to %s: %v", num, source, destination, err)
	}

	// Successfully copied, now delete the source file
	// err = os.Remove(source)
	// if err != nil {
	//     return fmt.Errorf("error deleting source file %s: %v", source, err)
	// }

	return nil
}
