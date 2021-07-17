package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

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

	api := operations.NewHelloAPIAPI(swaggerSpec)
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

	// Start server which listening
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}

//Health route returns OK
func Health(operations.CheckHealthParams) middleware.Responder {
	return operations.NewCheckHealthOK().WithPayload("OK")
}

//GetHelloUser returns Hello + your name
func GetGopherName(gopher operations.GetGopherNameParams) middleware.Responder {

	var URL string
	if gopher.Name != "" {
		URL = "https://github.com/scraly/gophers/raw/main/" + gopher.Name + ".png"
	} else {
		//by default we return Gandalf gopher
		URL = "https://github.com/scraly/gophers/raw/main/gandalf.png"
	}

	response, err := http.Get(URL)
	if err != nil {
		fmt.Println("error")
	}

	return operations.NewGetGopherNameOK().WithPayload(response.Body)
}

func GetGophers(operations.GetGophersParams) middleware.Responder {

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
		if *c.Name == ".gitignore" || *c.Name == "*.md" {
			continue
		}

		var name string = strings.Split(*c.Name, ".")[0]

		arr = append(arr, &models.Gopher{name, *c.Path, *c.DownloadURL})

		fmt.Println(c)

		//fmt.Println(arr)
	}

	return operations.NewGetGophersOK().WithPayload(arr)
}
