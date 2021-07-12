package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gaelleacas/kutego-api/pkg/swagger/server/restapi"
	"github.com/gaelleacas/kutego-api/pkg/swagger/server/restapi/operations"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
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

	api.GetHelloUserHandler = operations.GetHelloUserHandlerFunc(GetHelloUser)

	api.GetGophersNameHandler = operations.GetGophersNameHandlerFunc(GetGophersName)

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
func GetHelloUser(user operations.GetHelloUserParams) middleware.Responder {
	return operations.NewGetHelloUserOK().WithPayload("Hello " + user.User + "!")
}

//GetHelloUser returns Hello + your name
func GetGophersName(name operations.GetGophersNameParams) middleware.Responder {

	reqImg, err := http.Get("https://github.com/scraly/gophers/raw/main/dr-who.png")
	// if err != nil {
	// 	fmt.Println("error")
	// }

	if err != nil {
		fmt.Fprintf(res, "Error %d", err)
		return
	}
	buffer := make([]byte, reqImg.ContentLength)
	io.ReadFull(reqImg.Body, buffer)
	// res.Header().Set("Content-Length", fmt.Sprint(reqImg.ContentLength))
	// res.Header().Set("Content-Type", reqImg.Header.Get("Content-Type"))
	// res.Write(buffer)
	// req.Body.Close()

	return operations.NewGetGophersNameOK().WithPayload(buffer)
}
