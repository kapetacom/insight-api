package main

import (
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kapetacom/insight-api/handlers"
	kapetajwt "github.com/kapetacom/insight-api/jwt"
	"github.com/kapetacom/insight-api/middleware"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(mw.LoggerWithConfig(mw.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/healthz"
		},
	}))

	host := os.Getenv("JWT_PUBLIC_KEY_HOST")
	if host == "" {
		host = "http://localhost:5940"
	}
	log.Println("Using JWT public key host: " + host)
	config := echojwt.Config{
		// specify the function that returns the public key that will be used to verify the JWT
		KeyFunc: middleware.FetchKey(host + "/.well-known/jwks.json"),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &kapetajwt.KapetaClaims{}
		},
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
