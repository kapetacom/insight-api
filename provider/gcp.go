package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/kapetacom/insight-api/kubernetes"
	"github.com/kapetacom/insight-api/model"

	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

func GetDatabaseState(ctx context.Context) ([]model.ProviderState, error) {

	deployment, err := kubernetes.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := metadata.ProjectID()
	if err != nil {
		return nil, err
	}

	states := []model.ProviderState{}

	for _, svc := range deployment.Spec.Services {
		if svc.Kind == "kapeta/resource-type-postgresql" {
			// get status from cloudsql
			sqlService, err := sqladmin.NewService(ctx, option.WithScopes())
			if err != nil {
				return nil, err
			}
			splitName := strings.Split(deployment.Metadata.Name, "/")
			dbName := splitName[1] + "-" + svc.Id
			handle := splitName[0]
			log.Println("Trying to get status for CloudSQL database: ", dbName)
			// Retrieve the instance status
			list, err := sqlService.Instances.List(projectID).Context(ctx).Do()
			if err != nil {
				return nil, err
			}
			for _, db := range list.Items {
				if db.Settings.UserLabels["kapeta-deploymentname"] == dbName && db.Settings.UserLabels["kapeta-handle"] == handle {
					states = append(states, model.ProviderState{
						Name:  svc.Id,
						State: stateMapper(db.State),
					})

					fmt.Printf("Instance status: %s\n", db.State)
				}
			}

		}
	}
	return states, nil
}

func stateMapper(gcpState string) string {
	switch gcpState {
	case "RUNNABLE":
		return "Ready"
	case "PENDING_CREATE":
		return "Pending"
	case "MAINTENANCE":
		return "Maintenance"
	case "PENDING_DELETE":
		return "Pending"
	case "PENDING_MAINTENANCE":
		return "Pending"
	case "FAILED":
		return "Failed"
	case "UNKNOWN_STATE":
		return "Unknown"
	default:
		return "Unknown"
	}
}
