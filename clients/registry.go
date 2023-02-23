package clients

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/blockwarecom/insight-api/model"
	"github.com/pkg/errors"
)

type Client struct {
	RegistryServerURL string
	ServiceAccount    string
}

func (c *Client) GetRegistryCurrent(token string, handle string, name string) (*model.PublicAssetVersion, error) {
	req, err := http.NewRequest(http.MethodGet, c.RegistryServerURL+"/v1/registry/"+handle+"/"+name+"/current", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed to get current version from registry using '%v'", req.URL.String()))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed to get current version from registry using '%v', status code %v", req.URL.String(), resp.StatusCode))
	}
	version := &model.PublicAssetVersion{}
	err = json.NewDecoder(resp.Body).Decode(version)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed to decode current version from registry using '%v'", req.URL.String()))
	}
	return version, nil
}
