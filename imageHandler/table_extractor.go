package imagehandler

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Detecting tables in an image and returning their bounding boxes.
func DetectTable(img gocv.Mat) []image.Rectangle {
	contours := gocv.FindContours(img, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	var tableBoxes []image.Rectangle

	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		rectBound := gocv.BoundingRect(contour)
		aspectRatio := float64(rectBound.Dx()) / float64(rectBound.Dy())
		if aspectRatio > 2 && rectBound.Dy() > 20 { // Might need fine tuning. This is a condition for table like structures.
			tableBoxes = append(tableBoxes, rectBound)
		}
	}

	return tableBoxes
}

// Extracting table regions from an image and saving them as separate images.
func ExtractTableImages(img gocv.Mat, tableBoxes []image.Rectangle, outDir string) error {
	for i, rectBound := range tableBoxes {
		tableImage := img.Region(rectBound) // Returns a Mat with poointers to the region in the rectBound.
		outFilePath := fmt.Sprintf("%s/table_%d.jpg", outDir, i)
		gocv.IMWrite(outFilePath, tableImage)
		tableImage.Close()
	}

	return nil
}
