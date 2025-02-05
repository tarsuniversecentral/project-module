package service

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tarsuniversecentral/project-module/internal/dto"
	"github.com/tarsuniversecentral/project-module/pkg/utils"
)

type FileService struct {
}

func NewFileService() *FileService {
	return &FileService{}
}

// ProcessUploads saves the uploaded PDF and image files concurrently.
// If any error occurs, it deletes all the files that were saved.
const maxConcurrents = 10

func (fs *FileService) ProcessUploads(pdfHeaders, imageHeaders []*multipart.FileHeader) (dto.SavedFiles, error) {
	totalFiles := len(pdfHeaders) + len(imageHeaders)
	resultsCh := make(chan dto.FileResult, totalFiles)
	errCh := make(chan error, totalFiles)

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrents) // Semaphore for limiting concurrency

	// Helper function to save a file.
	saveFileConcurrently := func(header *multipart.FileHeader, fileType, destDir string) {
		defer wg.Done()
		defer func() { <-sem }()

		var allowedTypes []string
		if fileType == "pdf" {
			allowedTypes = []string{".pdf"}
		} else if fileType == "images" {
			allowedTypes = []string{".jpg", ".jpeg", ".png", ".svg"}
		}

		if !validateFileType(header, allowedTypes) {
			errCh <- fmt.Errorf("invalid file type for %s: %s", fileType, header.Filename)
			return
		}

		log.Printf("Saving %s file: %s", fileType, header.Filename)

		uniqueName, err := saveFile(header, destDir)
		if err != nil {
			errCh <- fmt.Errorf("error saving %s file %s: %w", fileType, header.Filename, err)
			return
		}

		log.Printf("Saved %s file: %s", fileType, uniqueName)

		resultsCh <- dto.FileResult{FileType: fileType, Filename: uniqueName}
	}

	// Process PDF files concurrently.
	for _, header := range pdfHeaders {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		go saveFileConcurrently(header, "pdf", "pdfs")
	}

	// Process image files concurrently.
	for _, header := range imageHeaders {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		go saveFileConcurrently(header, "images", "images")
	}

	wg.Wait()
	close(resultsCh)
	close(errCh)

	// Collect results.
	var savedFiles []dto.FileResult
	for res := range resultsCh {
		savedFiles = append(savedFiles, res)
	}

	// Check for errors.
	var errorsFound []error
	for err := range errCh {
		errorsFound = append(errorsFound, err)
	}

	// If there were any errors, delete all saved files concurrently.
	if len(errorsFound) > 0 {

		if err := fs.DeleteSavedFiles(savedFiles); err != nil {
			return dto.SavedFiles{}, fmt.Errorf("errors occurred while saving files: %v; errors occurred while deleting files: %v", errorsFound, err)
		}

		// Aggregate all errors into a single error message
		var errorMessages []string
		for _, err := range errorsFound {
			errorMessages = append(errorMessages, err.Error())
		}
		return dto.SavedFiles{}, fmt.Errorf("errors occurred while saving files: %v", strings.Join(errorMessages, "; "))
	}

	// Organize the results into the response struct.
	var response dto.SavedFiles
	for _, res := range savedFiles {
		if res.FileType == "pdf" {
			response.PDFFiles = append(response.PDFFiles, res.Filename)
		} else if res.FileType == "images" {
			response.ImageFiles = append(response.ImageFiles, res.Filename)
		}
	}

	return response, nil
}

func (fs *FileService) DeleteSavedFiles(savedFiles []dto.FileResult) error {
	sem := make(chan struct{}, maxConcurrents)
	errorCh := make(chan string, len(savedFiles)) // Buffered channel for error messages.

	var delWg sync.WaitGroup

	for _, res := range savedFiles {
		sem <- struct{}{}
		delWg.Add(1)

		go func(r dto.FileResult) {
			defer func() {
				<-sem
				delWg.Done()
			}()

			path := filepath.Join(r.FileType, r.Filename)
			if err := os.Remove(path); err != nil {
				log.Printf("Error deleting file %s: %v", path, err)
				errorCh <- fmt.Sprintf("deleting file %s: %v", path, err)
			}
		}(res)
	}

	// Wait for all operations to complete.
	go func() {
		delWg.Wait()
		close(errorCh)
	}()

	// Collect error messages.
	var errorMessages []string
	for msg := range errorCh {
		errorMessages = append(errorMessages, msg)
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("errors occurred while deleting files: %s", strings.Join(errorMessages, "; "))
	}

	return nil
}

func validateFileType(header *multipart.FileHeader, allowedTypes []string) bool {
	ext := filepath.Ext(header.Filename)
	for _, t := range allowedTypes {
		if strings.EqualFold(ext, t) {
			return true
		}
	}
	return false
}

// saveFile saves an individual file to the destination directory.
// It opens the uploaded file, creates a new file with a unique filename, and copies the content.
func saveFile(header *multipart.FileHeader, destDir string) (string, error) {

	if err := createDirIfNotExist(destDir); err != nil {
		return "", fmt.Errorf("creating directory %s: %w", destDir, err)
	}

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	uniqueName := utils.GenerateUniqueFilename(header.Filename)
	dstPath := filepath.Join(destDir, uniqueName)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("creating destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("copying file: %w", err)
	}
	return uniqueName, nil
}

// Function to create directories if they don't exist
func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// RetrieveFile retrieves a saved file based on its filename.
// It determines the correct directory by inspecting the file extension.
func (fs *FileService) RetrieveFile(filename string) (io.ReadCloser, error) {
	// Sanitize filename to prevent directory traversal attacks.
	sanitized := filepath.Base(filename)
	ext := filepath.Ext(sanitized)
	destDir, err := getDestinationDir(ext)
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(destDir, sanitized)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %q not found in directory %q", sanitized, destDir)
		}
		return nil, fmt.Errorf("error opening file %q: %w", filePath, err)
	}

	return file, nil
}

// getDestinationDir returns the destination directory based on the file extension.
func getDestinationDir(ext string) (string, error) {
	ext = strings.ToLower(ext)
	switch ext {
	case ".pdf":
		return "pdfs", nil
	case ".jpg", ".jpeg", ".png", ".svg":
		return "images", nil
	default:
		return "", fmt.Errorf("unsupported file extension %q", ext)
	}
}
