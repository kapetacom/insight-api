package logging

import (
	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kapetacom/insight-api/jwt"
	"github.com/kapetacom/insight-api/scopes"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"net/http"
	"strings"
)

func logClient(ctx context.Context) (*logadmin.Client, error) {
	// Read the service account file
	creds, err := google.FindDefaultCredentials(ctx, logging.ReadScope)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusForbidden, "failed to read service account file")
	}
	client, err := logadmin.NewClient(ctx, creds.ProjectID, option.WithCredentials(creds))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusForbidden, "failed to create log admin client")
	}
	return client, err
}

func GCPLogHandler(c echo.Context) error {
	// TODO: Verify current user has access proper to this deployment
	instanceId := c.Param("instance")
	deploymentHandle := c.Param("deploymentHandle")
	deploymentName := c.Param("deploymentName")

	if !jwt.HasScopeForHandle(c, deploymentHandle, scopes.LOGGING_READ_SCOPE) {
		return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("user does not have access to this deployment, missing scope %v for %v", scopes.LOGGING_READ_SCOPE, deploymentHandle))
	}

	// In labels "/" is not allowed - so it's seperated by "-" instead
	deployment := deploymentHandle + "-" + deploymentName

	client, err := logClient(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create log client")
	}

	filter := "labels.\"k8s-pod/instance\"=\"" + instanceId + "\" " +
		"labels.\"k8s-pod/deployment\"=\"" + deployment + "\" " +
		"resource.type=\"k8s_container\""

	it := client.Entries(c.Request().Context(), logadmin.Filter(filter), logadmin.NewestFirst())
	pageToken := ""

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	enc := json.NewEncoder(c.Response())
	var gcpLogEntries []*logging.Entry
	entries := make([]*LogEntry, 0)
	for {
		nextTok, err := iterator.NewPager(it, 100, pageToken).NextPage(&gcpLogEntries)
		if err != nil {
			// if context is cancelled, we can ignore the error
			if c.Request().Context().Err() != nil {
				return nil
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get next page of logs: %v", err))
		}
		for _, gcpLogEntry := range gcpLogEntries {
			logEntry := LogEntry{
				Entity:    gcpLogEntry.Resource.Labels["container_name"],
				Timestamp: gcpLogEntry.Timestamp.UnixMilli(),
				Severity:  strings.ToUpper(gcpLogEntry.Severity.String()),
				Message:   fmt.Sprintf("%v", gcpLogEntry.Payload),
			}
			entries = append(entries, &logEntry)
		}

		c.Response().Flush()
		if nextTok == "" {
			break
		}
		pageToken = nextTok
	}

	err = enc.Encode(entries)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to encode log output")
	}
	return nil
}
