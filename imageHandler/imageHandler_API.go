package imagehandler

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetTextFromImages(convertedPdfDir string) error {
	// Going over the pages (as jpegs) and preprocessing them.
	dir, err1 := os.Open(convertedPdfDir)
	files, err2 := os.ReadDir(convertedPdfDir)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("couldn't go over files in dir %s: %v", convertedPdfDir, err)
	}

	for file := range files {
		if filepath.Ext() == ".jpeg" {

		}
	}

	return nil
}
