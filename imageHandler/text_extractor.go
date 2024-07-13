package imagehandler

import (
	"os"

	gosseract "github.com/otiai10/gosseract/v2"
)

// Extracting text using Gosseract.
func ExtractTextFromImage(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage(imagePath)
	text, err := client.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}

func SaveExtractedText(text, outFilePath string) error {
	err := os.WriteFile(outFilePath, []byte(text), 0644)
	if err != nil {
		return err
	}

	return nil
}
