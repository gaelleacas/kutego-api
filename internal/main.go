package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/gaelleacas/kutego-api/pkg/swagger/server/models"
	"github.com/gaelleacas/kutego-api/pkg/swagger/server/restapi"
	"github.com/gaelleacas/kutego-api/pkg/swagger/server/restapi/operations"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/google/go-github/v37/github"
)

func main() {

	// Initialize Swagger
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewKutegoAPIAPI(swaggerSpec)
	// Use swaggerUI instead of reDoc on /docs
	api.UseSwaggerUI()

	server := restapi.NewServer(api)

	defer func() {
		if err := server.Shutdown(); err != nil {
			// error handle
			log.Fatalln(err)
		}
	}()

	server.Port = 8080

	api.CheckHealthHandler = operations.CheckHealthHandlerFunc(Health)

	api.GetGopherNameHandler = operations.GetGopherNameHandlerFunc(GetGopherName)

	api.GetGophersHandler = operations.GetGophersHandlerFunc(GetGophers)

	api.GetGopherRandomHandler = operations.GetGopherRandomHandlerFunc(GetGopherRandom)

	// Start server which listening
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}

//Health route returns OK
func Health(operations.CheckHealthParams) middleware.Responder {
	return operations.NewCheckHealthOK().WithPayload("OK")
}

//GetGopherName returns Gopher image (png)
func GetGopherName(gopher operations.GetGopherNameParams) middleware.Responder {

	var URL string
	if gopher.Name != "" {
		URL = "https://github.com/scraly/gophers/raw/main/" + gopher.Name + ".png"
	} else {
		//by default we return Gandalf gopher
		URL = "https://github.com/scraly/gophers/raw/main/fire-gopher.png"
	}

	response, err := http.Get(URL)
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		srcImage, _ := getFireGopherError("Ooops, Error")
		return operations.NewGetGopherNameOK().WithPayload(convertImgToIoCloser(srcImage))
	}

	srcImage, _, err := image.Decode(response.Body)
	if err != nil {
		log.Fatalf("failed to Decode image: %v", err)
	}

	if gopher.Size != nil {
		srcImage = resizeImage(srcImage, *gopher.Size)
	}

	return operations.NewGetGopherNameOK().WithPayload(convertImgToIoCloser(srcImage))
}

/**
Display Gopher list with optional filter
*/
func GetGophers(gopher operations.GetGophersParams) middleware.Responder {

	gophersList := GetGophersList()
	var arr []*models.Gopher
	for key, value := range gophersList {
		if value.Name == *gopher.Name {
			arr = append(arr, gophersList[key])
			return operations.NewGetGophersOK().WithPayload(arr)
		}
	}

	return operations.NewGetGophersOK().WithPayload(gophersList)
}

/**
Display a random Gopher Image
*/
func GetGopherRandom(gopher operations.GetGopherRandomParams) middleware.Responder {
	var URL string

	// Get Gophers List
	arr := GetGophersList()

	// Get a Random Index
	rand.Seed(time.Now().UnixNano())
	var index int
	index = rand.Intn(len(arr) - 1)

	URL = "https://github.com/scraly/gophers/raw/main/" + arr[index].Name + ".png"

	response, err := http.Get(URL)
	if err != nil {
		fmt.Println("error")
		srcImage, _ := getFireGopherError("Ooops, Error")
		return operations.NewGetGopherNameOK().WithPayload(convertImgToIoCloser(srcImage))
	}
	defer response.Body.Close()

	srcImage, _, err := image.Decode(response.Body)
	if err != nil {
		log.Fatalf("failed to Decode image: %v", err)
	}

	if gopher.Size != nil {
		srcImage = resizeImage(srcImage, *gopher.Size)
	}

	return operations.NewGetGopherNameOK().WithPayload(convertImgToIoCloser(srcImage))
}

/**
Display Fire Gopher with a message (error)
*/
func getFireGopherError(message string) (image.Image, error) {

	// Open local file
	file, err := os.Open("./assets/fire-gopher.png")
	if err != nil {
		log.Fatalf("failed to Open fire-gopher image: %v", err)
	}

	srcImage, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("failed to Decode image: %v", err)
		return srcImage, err
	}
	// Add Text on Gopher
	srcImage, err = TextOnGopher(srcImage, "Ooops, Error! It's on fire!")

	// Resize Image
	srcImage = resizeImage(srcImage, "medium")
	if err != nil {
		log.Fatalf("failed to put Text on Gopher: %v", err)
		return srcImage, err
	}

	return srcImage, nil
}

/**
Get Gophers List from Scraly repository
*/
func GetGophersList() []*models.Gopher {

	client := github.NewClient(nil)

	// list public repositories for org "github"
	ctx := context.Background()
	// list all repositories for the authenticated user
	_, directoryContent, _, err := client.Repositories.GetContents(ctx, "scraly", "gophers", "/", nil)
	if err != nil {
		fmt.Println(err)
	}

	var arr []*models.Gopher

	for _, c := range directoryContent {
		if *c.Name == ".gitignore" || *c.Name == "README.md" {
			continue
		}

		var name string = strings.Split(*c.Name, ".")[0]

		arr = append(arr, &models.Gopher{name, *c.Path, *c.DownloadURL})

	}

	return arr
}

/**
Resize Image
*/
func resizeImage(srcImage image.Image, size string) image.Image {

	var height int
	switch size {
	case "x-small":
		height = 50
	case "small":
		height = 100
	case "medium":
		height = 300
	default:
		// Mouhouhahaha!
		height = 1000
	}

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	srcImage = imaging.Resize(srcImage, 0, height, imaging.Lanczos)

	return srcImage

}

/**
Convert Image to io.close (for reply format)
*/
func convertImgToIoCloser(srcImage image.Image) io.ReadCloser {
	encoded := &bytes.Buffer{}
	png.Encode(encoded, srcImage)

	return ioutil.NopCloser(encoded)
}

/**
Add text on Image
*/
func TextOnGopher(bgImage image.Image, text string) (image.Image, error) {

	imgWidth := bgImage.Bounds().Dx()
	imgHeight := bgImage.Bounds().Dy()

	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(bgImage, 0, 0)

	if err := dc.LoadFontFace("assets/FiraSans-Light.ttf", 50); err != nil {
		return nil, err
	}

	x := float64((imgWidth / 2))
	y := float64((imgHeight / 12))
	maxWidth := float64(imgWidth) - 60.0
	dc.SetColor(color.Black)
	dc.DrawStringWrapped(text, x, y, 0.5, 0.5, maxWidth, 1.5, gg.AlignRight)

	return dc.Image(), nil
}
