package handlers

import (
	"encoding/base64"
	"os"

	"github.com/blockwarecom/insight-api/clients"
)

type Routes struct {
	Clients *clients.Client
}

func NewRoutes() *Routes {
	handlers := &Routes{
		Clients: &clients.Client{
			RegistryServerURL: "http://localhost:5941",
		},
	}

	if os.Getenv("REGISTRY_SERVER_URL") != "" {
		handlers.Clients.RegistryServerURL = os.Getenv("REGISTRY_SERVER_URL")
	}

	serviceAccount := os.Getenv("SERVICE_ACCOUNT")
	if serviceAccount == "" {
		panic("Environment SERVICE_ACCOUNT not found")
	}
	sa, err := base64.StdEncoding.DecodeString(serviceAccount)
	if err != nil {
		panic("Failed to decode service account")
	}
	handlers.Clients.ServiceAccount = string(sa)
	return handlers
}
