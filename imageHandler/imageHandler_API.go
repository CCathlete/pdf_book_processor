package imagehandler

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetTextFromImages(convertedPdfDir, outputTextDir string, chapterStartingPageNum int) error {
	// Going over the pages (as jpegs) and preprocessing them.
	dir, err1 := os.Open(convertedPdfDir)

	fileNames, err2 := dir.Readdirnames(-1)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("couldn't go over files in dir %s: %v\n%v", convertedPdfDir, err1, err2)
	}
	dir.Close()

	totalExtractionFailure := true

	for imgNum, fileName := range fileNames {
		fileExtension := filepath.Ext(fileName)
		if fileExtension == ".jpeg" {
			ExtractionFailure := false

			preProcessImgPath := fmt.Sprintf("%s/%s", convertedPdfDir, fileName)
			processedImgPath, err := processImage(preProcessImgPath, true)
			if err != nil {
				fmt.Printf("couldn't process image %s: %v", preProcessImgPath, err)
				processedImgPath = preProcessImgPath // We could still try to work with the pre processed image.
			}

			// Creating the path of the output dir for the inner images and tables for each page (jpeg) of the chapter.
			// The directory itself is created in extractInnerImages (if needed).
			pageNum := imgNum + chapterStartingPageNum
			outDir := fmt.Sprintf("%s/page_%d", convertedPdfDir, pageNum)

			// Loading the image as gocv.Mat
			processedImg, err := loadImage(processedImgPath)
			if err != nil {
				fmt.Printf("couldn't load processed image %s: %v", preProcessImgPath, err)
				// If the loading failed we can't work with this image.
			} else {
				// Detecting inner images and extracting them into new jpegs inside outDir we chose.
				innerImgBoundaries, err := detectInnerImages(processedImg)
				if err != nil {
					fmt.Printf("couldn't detect inner images in %s: %v", processedImgPath, err)
				}

				err = extractInnerImages(processedImg, innerImgBoundaries, outDir, pageNum)
				if err != nil {
					fmt.Printf("couldn't extract inner images in %s: %v", processedImgPath, err)
				}

				// Erasing the inner images from the preprocessed image.
				maskRegions(&processedImg, innerImgBoundaries)

				// Detecting inner tables and extracting them into new jpegs inside outDir.
				innerTableBoundaries := detectTable(processedImg)
				err = extractTableImages(processedImg, innerTableBoundaries, outDir)
				if err != nil {
					fmt.Printf("couldn't extract inner tables in %s: %v", processedImgPath, err)
				}

				// Erasing the inner tables from the preprocessed image.
				maskRegions(&processedImg, innerTableBoundaries)

				// Going over outDir for each page of the chapter, extracting the text from all of the images we extracted.
				dir, err1 := os.Open(outDir)

				fileNames, err2 := dir.Readdirnames(-1)
				if err1 != nil || err2 != nil {
					return fmt.Errorf("couldn't go over files in dir %s: %v\n%v", outDir, err1, err2)
				}
				dir.Close()

				for _, innerJpeg := range fileNames {
					innerJpegPath := fmt.Sprintf("%s/%s", outDir, innerJpeg)
					text, err := extractTextFromImage(innerJpegPath)
					if err != nil {
						fmt.Printf("couldn't extract text from %s: %v", innerJpegPath, err)
					} else if saveExtractedText(text, fmt.Sprintf("%s/book_content.txt", outputTextDir)) != nil {
						fmt.Printf("couldn't save the extracted text: %v", err)
						ExtractionFailure = true
					}
				}
			}
			totalExtractionFailure = totalExtractionFailure && ExtractionFailure
		}
	}

	if totalExtractionFailure {
		return fmt.Errorf("couldn't extract text at all")
	} else {
		return nil
	}
}
