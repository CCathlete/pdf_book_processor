package pdfhandler

import (
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"

	license "github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/model"
	render "github.com/unidoc/unipdf/v3/render"
)

func convertToImages(pdfPath, outputDirPath string) error {
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

		imagePath := filepath.Join(outputDirPath, fmt.Sprintf("page_%03d.jpg", pageNum))
		outImg, err := os.Create(imagePath)
		if err != nil {
			return fmt.Errorf("failed to create an image file for page %d: %v", pageNum, err)
		}
		defer outImg.Close()

		err = jpeg.Encode(outImg, img, &jpeg.Options{Quality: 100})
		if err != nil {
			return fmt.Errorf("failed to save image for page %d: %v", pageNum, err)
		}
	}

	// Setting the log level to warning.
	//common.SetLogger(common.NewConsoleLogger(common.LogLevelWarning))

	return nil
}
