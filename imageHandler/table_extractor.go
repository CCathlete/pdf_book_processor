package imagehandler

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Detecting tables in an image and returning their bounding boxes.
func detectTable(img gocv.Mat) []image.Rectangle {
	// Detection is based on the features of vertical and horizontal lines in tables.

	// Detecting horizontal lines.
	horizontal := gocv.NewMat()
	horizontalKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 30, Y: 1})
	gocv.MorphologyEx(img, &horizontal, gocv.MorphOpen, horizontalKernel)
	defer horizontal.Close()

	// Detecting vertical lines.
	vertical := gocv.NewMat()
	verticalKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 1, Y: 30})
	gocv.MorphologyEx(img, &vertical, gocv.MorphOpen, verticalKernel)
	defer vertical.Close()

	// Combining horizontal and vertical lines.
	bothLines := gocv.NewMat()
	gocv.AddWeighted(horizontal, 0.5, vertical, 0.5, 0, &bothLines)
	defer bothLines.Close()

	// Finding the contours of each table.
	contours := gocv.FindContours(bothLines, gocv.RetrievalExternal, gocv.ChainApproxSimple)
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
func extractTableImages(img gocv.Mat, tableBoxes []image.Rectangle, outDir string) error {
	for i, rectBound := range tableBoxes {
		tableImage := img.Region(rectBound) // Returns a Mat with poointers to the region in the rectBound.
		outFilePath := fmt.Sprintf("%s/table_%d.jpg", outDir, i)
		gocv.IMWrite(outFilePath, tableImage)
		tableImage.Close()
	}

	return nil
}
