package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type LogEntry struct {
	Entity    string `json:"entity"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

func logClient(ctx context.Context, serviceAccountFile []byte) (*logadmin.Client, error) {
	// Read the service account file
	creds, err := google.CredentialsFromJSON(ctx, serviceAccountFile, logging.ReadScope)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusForbidden, "failed to read service account file")
	}
	client, err := logadmin.NewClient(ctx, creds.ProjectID, option.WithCredentials(creds))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusForbidden, "failed to create log admin client")
	}
	return client, err
}

func (h *Routes) LogHandler(c echo.Context) error {
	handle := c.Param("handle")
	environment := c.Param("environment")
	blockName := c.Param("block")

	client, err := logClient(c.Request().Context(), []byte(h.Clients.ServiceAccount))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create log client")
	}

	clustername := handle + "-" + environment
	filter := "resource.labels.container_name=\"" + blockName + "\" resource.labels.cluster_name=\"" + clustername + "\" resource.type=\"k8s_container\" AND log_id(\"stderr\") AND severity>=ERROR resource.labels.namespace_name=\"default\""

	it := client.Entries(c.Request().Context(), logadmin.Filter(filter), logadmin.NewestFirst())
	pageToken := ""

	enc := json.NewEncoder(c.Response())
	var entries []*logging.Entry

	for {
		nextTok, err := iterator.NewPager(it, 100, pageToken).NextPage(&entries)
		if err != nil {
			// if contenxt is cancelled, we can ignore the error
			if c.Request().Context().Err() != nil {
				return nil
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get next page of logs: %v", err))
		}
		for _, entry := range entries {
			le := LogEntry{
				Entity:    entry.Resource.Labels["container_name"],
				Timestamp: entry.Timestamp.Format(time.StampMicro),
				Severity:  entry.Severity.String(),
				Message:   fmt.Sprintf("%v", entry.Payload),
			}
			err = enc.Encode(le)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to encode log entry")
			}
		}
		c.Response().Flush()
		if nextTok == "" {
			break
		}
		pageToken = nextTok
	}
	return nil
}