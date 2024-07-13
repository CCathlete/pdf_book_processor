package pdfhandler

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	license "github.com/unidoc/unipdf/v3/common/license"
	model "github.com/unidoc/unipdf/v3/model"
	render "github.com/unidoc/unipdf/v3/render"
)

func convertToImagesUniPdf(pdfPath, outputDirPath string) error {
	// Creating the output dir if it doesn't exist. If it exists mkdir will get an error but it's ok for us because the dir exists.
	// Note: I used isNotExists because it takes into account an empty file.
	if err := os.MkdirAll(outputDirPath, 0755); os.IsNotExist(err) {
		// If the dir doesn't exist but we still got an error.
		return fmt.Errorf("error when trying to create the output dir even"+
			"though it doesn't already exist: %v", err)
	}

	err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
	if err != nil {
		fmt.Printf("\nProblem setting the license key for UniPDF: %v\n", err)
		return err
	}

	file, err := os.Open(pdfPath)
	if err != nil {
		fmt.Printf("\nProblem opening the pdf at %s: %v\n", pdfPath, err)
		return err
	}
	defer file.Close()

	pdfReader, err := model.NewPdfReader(file)
	if err != nil {
		return fmt.Errorf("failed to create pdf reader: %v", err)
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return fmt.Errorf("failed to check encryption for the pdf at %s: %v", pdfPath, err)
	}
	if isEncrypted {
		return fmt.Errorf("the pdf is encrypted and cannot be processed. please decrypt first")
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return fmt.Errorf("failed to get the number of pages: %v", err)
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return fmt.Errorf("failed to get page %d: %v", pageNum, err)
		}

		pageRenderer := render.NewImageDevice()
		img, err := pageRenderer.Render(page)
		if err != nil {
			return fmt.Errorf("failed to render page %d: %v", pageNum, err)
		}
		scaledImage := scaleImage(img, img.Bounds().Dx()*7, img.Bounds().Dy()*7)

		imagePath := filepath.Join(outputDirPath, fmt.Sprintf("page_%03d.jpg", pageNum))
		outImg, err := os.Create(imagePath)
		if err != nil {
			return fmt.Errorf("failed to create an image file for page %d: %v", pageNum, err)
		}
		defer outImg.Close()

		err = jpeg.Encode(outImg, scaledImage, &jpeg.Options{Quality: 100})
		if err != nil {
			return fmt.Errorf("failed to save image for page %d: %v", pageNum, err)
		}
	}

	// Setting the log level to warning.
	//common.SetLogger(common.NewConsoleLogger(common.LogLevelWarning))

	return nil
}

func scaleImage(img image.Image, width, height int) image.Image {
	// Creating a new image with the desired size.
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Scaling the original image by transforming it to the new image.
	newBounds := newImg.Bounds()
	newDx, newDy := newBounds.Dx(), newBounds.Dy()
	for y := 0; y < newDy; y++ {
		for x := 0; x < newDx; x++ {
			// Matching every x, y in the new image to an x, y in the original image, using the aspect ratio.
			originalX := x * img.Bounds().Dx() / newDx
			originalY := y * img.Bounds().Dy() / newDy

			// Setting the values of the pixels in the new image correspondingly.
			newImg.Set(x, y, img.At(originalX, originalY))
		}
	}

	return newImg
}
