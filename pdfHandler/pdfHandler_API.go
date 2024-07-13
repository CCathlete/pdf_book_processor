package pdfhandler

import (
	"fmt"
	"log"
)

func ConvertPdfToImages(pdfPath, outDirPath string) {
	err := convertToImages(pdfPath, outDirPath)
	if err != nil {
		log.Fatalf("ConvertPdfToImages: %v", err)
	}

	fmt.Printf("\nPDF converted to images successfully at %s\n.", outDirPath)
}
