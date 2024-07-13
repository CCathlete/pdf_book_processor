package pdfhandler

import (
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	license "github.com/unidoc/unipdf/v3/common/license"
	extractor "github.com/unidoc/unipdf/v3/extractor"
	model "github.com/unidoc/unipdf/v3/model"
	"gocv.io/x/gocv"
)

func ConvertToTextWhenNotScanned(pdfPath string) {
	cmd := exec.Command("bash", "-c", "pdftotext "+pdfPath)
	stdOutErr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(stdOutErr)
		log.Fatalf("Failed to run pdftotext using the path %s, the error was: %v\n", pdfPath, err)
	}
}

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

func runTesseractOCR(imagepath string) string {
	tempOutFile := "output" // Tesseract automatically adds .txt
	cmd := exec.Command("/bin/tesseract", imagepath, tempOutFile)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to convert image with tesseract %v\n", err)
		return ""
	}

	text, err := os.ReadFile(tempOutFile + ".txt")
	if err != nil {
		log.Fatalf("Failed to convert image with tesseract %v\n", err)
		return ""
	}

	os.Remove(tempOutFile + ".txt")

	return string(text)
}

func ImagesToText(inputDir, pdfPath string, needProcessing bool) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v\n", inputDir, err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".jpeg" {
			imagePath := filepath.Join(inputDir, file.Name())

			fmt.Printf("Processing image %s.\n", imagePath)
			processedImagePath := ProcessImage(imagePath, needProcessing)

			fmt.Printf("Using tesseract on image %s.\n", processedImagePath)

			text := runTesseractOCR(processedImagePath)

			// Adding the text to the txt file.
			txtPath := strings.Replace(pdfPath, "pdf", "txt", -1) // -1 means all instances.
			file, err := os.OpenFile(txtPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
			if err != nil {
				log.Fatalf("Failed to a file in path %s, the error was: %v\n", txtPath, err)
			}
			defer file.Close()

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

func ConvertWithUniPDF(pdfPath, outputDirPath string) {
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

			text := runTesseractOCR(imagePath)
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
