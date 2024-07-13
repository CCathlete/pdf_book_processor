package imagehandler

import (
	"fmt"
	"os"
	"path/filepath"

	"gocv.io/x/gocv"
)

func ProcessImage(imagePath string, needProcessing bool) (outPath string) {
	outDir := fmt.Sprintf("%s/AfterProcessing", filepath.Dir(imagePath))
	outPath = fmt.Sprintf("%s/Processed_%s", outDir, filepath.Base(imagePath))

	if needProcessing {
		// Converting to grayscale.
		image := gocv.IMRead(imagePath, gocv.IMReadGrayScale)
		if image.Empty() {
			fmt.Printf("Error when reading image %s\n", imagePath)
			return ""
		}
		defer image.Close()

		// Applying adaptive thresholding.
		afterThresh := gocv.NewMat()
		defer afterThresh.Close()
		gocv.AdaptiveThreshold(
			image,
			&afterThresh,
			255,
			gocv.AdaptiveThresholdMean,
			gocv.ThresholdBinary,
			11,
			2,
		)

		// Creating the output dir if it doesn't exist.
		// Note: I used isNotExists because it takes into account an empty file.
		if err := os.Mkdir(outDir, 0755); os.IsNotExist(err) {
			// If the dir doesn't exist but we still got an error.
			fmt.Printf("Error when trying to create the output dir even"+
				"though it doesn't already exist: %v.\n", err)
		}

		// Saving the processed image.
		if imageWasWritten := gocv.IMWrite(outPath, afterThresh); !imageWasWritten {
			fmt.Printf("Error when writing the processed image: %s\n", outPath)
			return ""
		}

		fmt.Printf("Successfully processed and saved the image to: %s\n", outPath)
	}

	return outPath
}
