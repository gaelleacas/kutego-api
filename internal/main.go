package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/disintegration/imaging"
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

	if response.StatusCode != 200 {
		URL = "https://github.com/scraly/gophers/raw/main/fire-gopher.png"
		response, err = http.Get(URL)
	}

	outputImg := response.Body

	if gopher.Size != nil {
		outputImg = resizeImage(*response, *gopher.Size)
	}

	return operations.NewGetGopherNameOK().WithPayload(outputImg)
}

func GetGophers(operations.GetGophersParams) middleware.Responder {

	arr := GetGophersList()

	return operations.NewGetGophersOK().WithPayload(arr)
}

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
	}

	outputImg := response.Body
	if gopher.Size != nil {
		outputImg = resizeImage(*response, *gopher.Size)
	}

	return operations.NewGetGopherNameOK().WithPayload(outputImg)
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

func resizeImage(response http.Response, size string) io.ReadCloser {
	srcImage, _, err := image.Decode(response.Body)

	if err != nil {
		log.Fatalf("failed to Decode image: %v", err)
	}

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

	//fmt.Println(src)
	encoded := &bytes.Buffer{}
	err = png.Encode(encoded, srcImage)

	return ioutil.NopCloser(encoded)

}
