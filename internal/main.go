package main

import (
	"fmt"
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

	response, err := http.Get("https://github.com/scraly/gophers/raw/main/dr-who.png")
	if err != nil {
		fmt.Println("error")
	}

	// buffer := make([]byte, response.ContentLength)
	// response.Body.Read(buffer)
	// buffer := make([]byte, response.ContentLength)
	// io.ReadFull(response.Body, buffer)
	// check := func(err error) {
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// check(err)
	// body, err := ioutil.ReadAll(response.Body)
	// check(err)

	// log.Printf("received: %d %s\n", response.StatusCode, body)

	// middleware.ResponderFunc(func(rw http.ResponseWriter, pr runtime.Producer) {
	// 	rw.Header().Add("Content-Length", fmt.Sprint(response.ContentLength))
	// 	rw.Header().Add("Content-Type", response.Header.Get("Content-Type"))

	// 	rw.WriteHeader(200)
	// 	rw.Write(buffer)
	// })
	return operations.NewGetGophersNameOK().WithPayload(response.Body)

}
