package imagehandler

import (
	"fmt"
	"image"
	"os"

	"gocv.io/x/gocv"
)

func detectInnerImages(img gocv.Mat) ([]image.Rectangle, error) {
	contours := gocv.FindContours(img, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	var imageBoxes []image.Rectangle

	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		rectBound := gocv.BoundingRect(contour)
		aspectRatio := float64(rectBound.Dx()) / float64(rectBound.Dy())
		if aspectRatio > 0.5 && aspectRatio < 2 && rectBound.Dy() > 20 {
			imageBoxes = append(imageBoxes, rectBound)
		}
	}

	return imageBoxes, nil
}

// Extracting the regions where there are images from the main image, as gocv.Mat, and saves them as separate images.
func extractInnerImages(img gocv.Mat, imageBoxes []image.Rectangle, outDir string, pageNum int) error {

	// Creating the output dir if it doesn't exist. If it exists mkdir will get an error but it's ok for us because the dir exists.
	// Note: I used isNotExists because it takes into account an empty file.
	if err := os.MkdirAll(outDir, 0755); os.IsNotExist(err) {
		// If the dir doesn't exist but we still got an error.
		return fmt.Errorf("error when trying to create the output dir even"+
			"though it doesn't already exist: %v", err)
	}
	for i, rectBound := range imageBoxes {
		croppedImg := img.Region(rectBound)

		croppedFilePath := fmt.Sprintf("%s/page_%d_image_%d.jpg", outDir, pageNum, i)
		croppedFileObject, err := os.Create(croppedFilePath)
		if err != nil {
			return fmt.Errorf("failed to create a cropped image in %s: %v", croppedFilePath, err)
		}
		defer croppedFileObject.Close()

		if ok := gocv.IMWrite(croppedFilePath, croppedImg); !ok {
			return fmt.Errorf("failed to write image %s to a jpg file: %v", croppedFilePath, err)
		}
	}

	return nil
}
