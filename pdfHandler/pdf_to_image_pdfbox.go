package pdfhandler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func convertToImgGhostscript(pdfPath, outDirPath, outputFormat string, dpi int) error {

	// Creating the output dir if it doesn't exist. If it exists mkdir will get an error but it's ok for us because the dir exists.
	// Note: I used isNotExists because it takes into account an empty file.
	if err := os.MkdirAll(outDirPath, 0755); os.IsNotExist(err) {
		// If the dir doesn't exist but we still got an error.
		return fmt.Errorf("error when trying to create the output dir even"+
			"though it doesn't already exist: %v", err)
	}
	outFilePattern := filepath.Join(outDirPath, "page-%03d."+outputFormat)
	cmd := exec.Command(
		"gs",
		"-dNOPAUSE",
		"-dBATCH",
		"-sDEVICE="+outputFormat,
		fmt.Sprintf("-r%d", dpi),
		"-sOutputFile="+outFilePattern,
		pdfPath,
	)

	outputByteSlice, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to convert pdf to jpgs: %v. output: %s", err, string(outputByteSlice))
	}

	fmt.Println("PDF pages converted to images successfully.")
	fmt.Printf("Conversion output: %s\n", string(outputByteSlice))

	return nil
}
