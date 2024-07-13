package main

import (
	pdfH "pdf_book_processor/pdfHandler"
)

func main() {
	pdfPath := "Chapter8_Pbook2.pdf"
	convertedImgsPath := "pdf_as_jpgs"
	pdfH.ConvertPdfToImages(pdfPath, convertedImgsPath)
}
