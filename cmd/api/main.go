// Package main starts the example Go Lambda
package main

import (
	awsLambda "github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/wimspaargaren/go-lambda-localstack-example/internal/api"
	"github.com/wimspaargaren/go-lambda-localstack-example/internal/lambda"
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	apiHandler := api.NewHandler()

	e.GET("/hello-world", apiHandler.HelloWorld)
	e.POST("/your-name", apiHandler.YourName)

	lambdaEchoHandler := lambda.NewEchoHandler(echoadapter.New(e))
	awsLambda.Start(lambdaEchoHandler.Handle)
}
