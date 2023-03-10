package main

import (
	"net/http"
	"os"

	"github.com/blockwarecom/insight-api/handlers"
	"github.com/blockwarecom/insight-api/middleware"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(mw.Logger())

	host := os.Getenv("JWT_PUBLIC_KEY_HOST")
	if host == "" {
		host = "http://localhost:5940"
	}

	config := echojwt.Config{
		// specify the function that returns the public key that will be used to verify the JWT
		KeyFunc: middleware.FetchKey(host + "/.well-known/jwks.json"),
	}

	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK!")
	})

	// Create a restricted group of routes that requires a valid JWT
	v1 := e.Group("/v1")
	v1.Use(echojwt.WithConfig(config))
	v1.Use(middleware.Restricted())

	routes := handlers.NewRoutes()

	v1.GET("/instances/:deploymentHandle/:deploymentName/:instance/logs", routes.LogHandler)
	// The :handle and :environment aren't really used in this route, but they are required to match the API of the local cluster service
	v1.GET("/status", routes.GetEnvironmentStatus)
	// Start the service and log if the server fails to start/crashes
	e.Logger.Fatal(e.Start(":1323"))
}
