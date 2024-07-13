package imagehandler

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

func DetectImages(img gocv.Mat) []image.Rectangle {
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

	return imageBoxes
}

// Extracting the regions where there are images from the main image, as gocv.Mat, and saves them as separate images.
func ExtractImageRegions(img gocv.Mat, imageBoxes []image.Rectangle, outDir string) error {
	for i, rectBound := range imageBoxes {
		imageRegion := img.Region(rectBound)
		outFilePath := fmt.Sprintf("%s/image_%d.jpg", outDir, i)
		gocv.IMWrite(outFilePath, imageRegion)
		imageRegion.Close()
	}

	return nil
}
