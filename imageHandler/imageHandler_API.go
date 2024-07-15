package imagehandler

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetTextFromImages(convertedPdfDir string, chapterStartingPageNum int) error {
	// Going over the pages (as jpegs) and preprocessing them.
	dir, err1 := os.Open(convertedPdfDir)

	fileNames, err2 := dir.Readdirnames(-1)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("couldn't go over files in dir %s: %v\n%v", convertedPdfDir, err1, err2)
	}
	dir.Close()

	for imgNum, fileName := range fileNames {
		if filepath.Ext(fileName) == ".jpeg" {
			imgProcessingFailed, imgLoadingFailed, imgDetectionFailed, imgExtractionFailed := false, false, false, false
			tableExtractionFailed := false

			preProcessImgPath := fmt.Sprintf("%s/%s", convertedPdfDir, fileName)
			processedImgPath, err := processImage(preProcessImgPath, true)
			if err != nil {
				fmt.Printf("couldn't process image %s: %v", preProcessImgPath, err)
				imgProcessingFailed = true
				processedImgPath = preProcessImgPath // We could still try to work with the pre processed image.
			}

			// Creating an output dir for the inner images and tables for each page (jpeg) of the chapter.
			outDir := fmt.Sprintf("%s/inner_images", processedImgPath)
			pageNum := imgNum + chapterStartingPageNum

			// Loading the image as gocv.Mat
			processedImg, err := loadImage(processedImgPath)
			if err != nil {
				fmt.Printf("couldn't load processed image %s: %v", preProcessImgPath, err)
				// If the loading failed we can't work with this image.
				imgLoadingFailed, imgDetectionFailed, imgDetectionFailed = true, true, true
				tableExtractionFailed = true
			}

			if !imgLoadingFailed {
				// Detecting inner images and extracting them into new jpegs inside outDir we chose.
				innerImgBoundaries, err := detectInnerImages(processedImg)
				if err != nil {
					fmt.Printf("couldn't detect inner images in %s: %v", processedImgPath, err)
					imgDetectionFailed = true
				}

				err = extractInnerImages(processedImg, innerImgBoundaries, outDir, pageNum)
				if err != nil {
					fmt.Printf("couldn't extract inner images in %s: %v", processedImgPath, err)
					imgExtractionFailed = true
				}

				// Detecting inner tables and extracting them into new jpegs inside outDir.
				innerTableBoundaries := detectTable(processedImg)
				err = extractTableImages(processedImg, innerTableBoundaries, outDir)
				if err != nil {
					fmt.Printf("couldn't extract inner tables in %s: %v", processedImgPath, err)
					tableExtractionFailed = true
				}
			}
		}
	}

	return nil
}
