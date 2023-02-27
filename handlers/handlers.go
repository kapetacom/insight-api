package handlers

import (
	"encoding/base64"
	"os"
)

type Client struct {
	RegistryServerURL string
	ServiceAccount    string
}

type Routes struct {
	Clients *Client
}

func NewRoutes() *Routes {
	handlers := &Routes{
		Clients: &Client{},
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
