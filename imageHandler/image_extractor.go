package imagehandler

import (
	"fmt"
	"image"
	"os"

	"gocv.io/x/gocv"
)

func isRegion(imgSection gocv.Mat) bool {
	stdDev, meanMat := gocv.NewMat(), gocv.NewMat()
	defer func() {
		stdDev.Close()
		meanMat.Close()
	}()
	gocv.MeanStdDev(imgSection, &meanMat, &stdDev)

	// meanMat and stdDev are vectors (1 x n channels matrix) to return values for each channel in RGB. In grayscale it's a 1 x 1 matrix.
	// So actually it only has a value in (0, 0).
	stdDevVal := stdDev.GetDoubleAt(0, 0)
	// Adjusting for a case we have multiple channels.
	if imgSection.Channels() > 1 {
		stdDevVal = (stdDev.GetDoubleAt(0, 0) + stdDev.GetDoubleAt(0, 1) + stdDev.GetDoubleAt(0, 2)) / 3
	}

	return stdDevVal > 35 // This is the threshold value for knowing if there's a sharp chage in brightness.
}

func detectInnerImages(img gocv.Mat) ([]image.Rectangle, error) {
	blur := gocv.NewMat()
	gocv.GaussianBlur(img, &blur, image.Point{X: 5, Y: 5}, 0, 0, gocv.BorderDefault)
	defer blur.Close()

	edges := gocv.NewMat()
	gocv.Canny(blur, &edges, 100, 200)
	defer edges.Close()

	contours := gocv.FindContours(edges, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	var imageBoxes []image.Rectangle

	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		rectBound := gocv.BoundingRect(contour)
		// Filtering out small contours.
		area := rectBound.Dx() * rectBound.Dy()
		aspectRatio := float64(rectBound.Dx()) / float64(rectBound.Dy())
		if area > 1000 && aspectRatio > 0.5 && aspectRatio < 2.0 {
			if isRegion(img.Region(rectBound)) {
				imageBoxes = append(imageBoxes, rectBound)
			}
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
