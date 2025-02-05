package dto

type SavedFiles struct {
	ImageFiles []string
	PDFFiles   []string
}

type FileResult struct {
	FileType string
	Filename string
}

// ConstructFileResults converts a SavedFiles instance into a slice of FileResult.
func ConstructFileResults(savedFiles SavedFiles) []FileResult {
	var fileResults []FileResult

	// Process image files
	for _, file := range savedFiles.ImageFiles {
		fileResults = append(fileResults, FileResult{
			FileType: "images",
			Filename: file,
		})
	}

	// Process PDF files
	for _, file := range savedFiles.PDFFiles {
		fileResults = append(fileResults, FileResult{
			FileType: "pdfs",
			Filename: file,
		})
	}

	return fileResults
}
