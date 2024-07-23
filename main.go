package main

import (
	"log"
	imgH "pdf_book_processor/imageHandler"
)

func main() {
	// pdfPath := "Chapter8_Pbook2.pdf"
	// convertedImgsPath := "pdf_as_jpgs"
	// pdfH.ConvertPdfToImages(pdfPath, convertedImgsPath)
	err := imgH.GetTextFromImages("./pdf_as_jpgs", ".", 343)
	if err != nil {
		log.Fatalf("Error in extraction of text: %v", err)
	}
}
