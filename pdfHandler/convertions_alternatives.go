package pdfhandler

import (
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	license "github.com/unidoc/unipdf/v3/common/license"
	extractor "github.com/unidoc/unipdf/v3/extractor"
	model "github.com/unidoc/unipdf/v3/model"

	imgH "pdf_book_processor/imageHandler"
)

func convertToTextWhenNotScanned(pdfPath string) {
	cmd := exec.Command("bash", "-c", "pdftotext "+pdfPath)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(stdOutErr)
		log.Fatalf("Failed to run pdftotext using the path %s, the error was: %v\n", pdfPath, err)
	}
}

func splitTextByAnimals(text string, animals []string) map[string]string {
	splittedText := map[string]string{}
	pattern := ""

	for i, animal := range animals {
		if i < len(animals)-1 {
			nextAnimal := animals[i+1]
			// Note: for extra security it's better to use regexp.QuteMeta(animal).
			// it makes sure to treat the animal names as raw strings with no special characters.
			pattern = fmt.Sprintf(`Parasites of %s(.*)Parsites of %s`, animal, nextAnimal)
		} else {
			// There is no next animal for the last animal.
			pattern = fmt.Sprintf(`Parasites of %s(.*)`, animal)
		}
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(text)
		if len(match) > 1 {
			if match[1] != "" {
				splittedText[animal] = match[1]
			}
		}
	}

	return splittedText
}

func convertWithUniPDF(pdfPath, outputDirPath string) {
	err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
	if err != nil {
		panic(err)
	}
	f, err := os.Open(pdfPath)
	if err != nil {
		log.Fatalf("Failed to open the file in path %s, the error was: %v\n", pdfPath, err)
	}

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		log.Fatalf("Failed to createnannew pdf reader for path %s, the error was: %v\n", pdfPath, err)
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		log.Fatalf("Failed to get number of pages, the error was: %v\n", err)
	}

	fmt.Printf("--------------------\n")
	fmt.Printf("PDF to text extraction:\n")
	fmt.Printf("--------------------\n")
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			log.Fatalf("Failed to get a page pbject for page num %d, the error was: %v\n", pageNum, err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Fatalf("Failed to create a new extractor for page num %d, the error was: %v\n", pageNum, err)
		}

		images, err := ex.ExtractPageImages(nil)
		if err != nil {
			log.Fatalf("Failed to extract an image from page num %d, the error was: %v\n",
				pageNum, err)
		}

		for j, image := range images.Images {
			// Extracting text from the am image with OCR.
			imagePath := fmt.Sprintf("%s/page_%d_image_%d.jpg", outputDirPath, i, j)
			imageFile, err := os.Create(imagePath)
			if err != nil {
				log.Fatalf("Failed to create a new extractor for page num %d, the error was: %v\n", pageNum, err)
			}
			defer imageFile.Close()

			goImage, err := image.Image.ToGoImage()
			if err != nil {
				log.Fatalf("Failed to convert to Go image %d on page %d: %v", j, i, err)
			}

			err = jpeg.Encode(imageFile, goImage, nil)
			if err != nil {
				log.Fatalf("Failed to encode image %d on page %d: %v", j, i, err)
			}

			// Adding the text to the txt file.
			txtPath := strings.Replace(pdfPath, "pdf", "txt", -1) // -1 means all instances.
			file, err := os.OpenFile(txtPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
			if err != nil {
				log.Fatalf("Failed to a file in path %s, the error was: %v\n", txtPath, err)
			}
			defer file.Close()

			text, err := imgH.ExtractTextFromImage(imagePath)
			if err != nil {
				log.Fatalf("Failed converting image %d to text: %v", j, err)
			}

			numOfBits, err := file.WriteString(text)
			if err != nil {
				log.Fatalf("Failed to write the text into a txt file, the error was: %v\n", err)
			} else {
				fmt.Printf("%d bits were written.\n", numOfBits)
			}

			if err := file.Sync(); err != nil {
				log.Fatalf("Failed to sync file: %v", err)
			}
		}

		// fmt.Println("------------------------------")
		// fmt.Printf("Page %d:\n", pageNum)
		// fmt.Printf("\"%s\"\n", text)
		// fmt.Println("------------------------------")
	}
}
