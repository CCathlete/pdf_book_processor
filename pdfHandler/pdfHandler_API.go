package pdfhandler

import (
	"fmt"
	"log"
)

func ConvertPdfToImages(pdfPath, outDirPath string) {
	err := convertToImgGhostscript(pdfPath, outDirPath, "jpeg", 300)
	if err != nil {
		log.Fatalf("convertToImgGhostscript: %v", err)
	}

	fmt.Printf("\nPDF converted to images successfully at %s\n.", outDirPath)
}
