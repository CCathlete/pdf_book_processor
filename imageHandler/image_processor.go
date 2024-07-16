package imagehandler

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"gocv.io/x/gocv"
)

func loadImage(imagePath string) (gocv.Mat, error) {
	image := gocv.IMRead(imagePath, gocv.IMReadAnyColor)
	if image.Empty() {
		return image, fmt.Errorf("error reading image from: %s", imagePath)
	}
	return image, nil
}

func convertToGray(image gocv.Mat) gocv.Mat {
	grayImage := gocv.NewMat()
	gocv.CvtColor(image, &grayImage, gocv.ColorBGRToGray)
	return grayImage
}

func thresholdImage(image gocv.Mat) gocv.Mat {
	binaryImage := gocv.NewMat()
	gocv.Threshold(
		image,
		&binaryImage,
		0,
		255,
		gocv.ThresholdBinary+gocv.ThresholdOtsu,
	)
	return binaryImage
}

func processImage(imagePath string, needProcessing bool) (string, error) {
	outDir := fmt.Sprintf("%s/AfterProcessing", filepath.Dir(imagePath))
	outPath := fmt.Sprintf("%s/Processed_%s", outDir, filepath.Base(imagePath))

	// Creating the output dir if it doesn't exist. If it exists mkdir will get an error but it's ok for us because the dir exists.
	// Note: I used isNotExists because it takes into account an empty file.
	if err := os.MkdirAll(outDir, 0755); os.IsNotExist(err) {
		// If the dir doesn't exist but we still got an error.
		return "", fmt.Errorf("error when trying to create the output dir even"+
			"though it doesn't already exist: %v", err)
	}

	if needProcessing {
		// Loading the image.
		beforeProcImage, err := loadImage(imagePath)
		if err != nil {
			return "", fmt.Errorf("\nProblem with loading the image: %v", err)
		}

		// Converting to grayscale.
		grayImage := convertToGray(beforeProcImage)
		if grayImage.Empty() {
			fmt.Printf("Error when converting to grayscale the image at %s\n", imagePath)
		}
		defer grayImage.Close()

		// Applying thresholding.
		afterThresh := thresholdImage(grayImage)
		defer afterThresh.Close()

		// Saving the processed image.
		if imageWasWritten := gocv.IMWrite(outPath, afterThresh); !imageWasWritten {
			fmt.Printf("Error when writing the processed image: %s\n", outPath)
		}

		fmt.Printf("Successfully processed and saved the image to: %s\n", outPath)
	}

	return outPath, nil
}

func maskRegions(img *gocv.Mat, regions []image.Rectangle) {
	for _, region := range regions {
		// Drawing a white rectangle over each region.
		gocv.Rectangle(img, region, color.RGBA{255, 255, 255, 0}, -1)
	}
}
