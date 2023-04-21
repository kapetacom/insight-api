package operators

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetDatabaseState(ctx context.Context) ([]model.OperatorState, error) {

	deployment, err := kubernetes.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := metadata.ProjectID()
	if err != nil {
		return nil, err
	}

	states := []model.OperatorState{}

	for _, svc := range deployment.Spec.Services {
		switch svc.Kind {
		case "kapeta/resource-type-postgresql":
			states = append(states, getPostgresState(ctx, deployment.Metadata.Name, svc.Id, projectID))
		case "kapeta/resource-type-mongodb":
			states = append(states, getMongoDBState(svc.Id))
		}
	}

	return states, nil
}

func getMongoDBState(id string) model.OperatorState {
	state := model.OperatorState{
		Name:  id,
		State: "Failed",
	}
	clientset, err := kubernetes.KubernetesClient()
	if err != nil {
		log.Printf("error getting kubernetes client: %v\n", err)
		return state
	}
	mongoDeployment, err := clientset.AppsV1().Deployments("infrastructure").Get(context.Background(), id, metav1.GetOptions{})
	if err != nil {
		log.Printf("error getting deployment: %v\n", err)
		return state
	}
	readyReplicas := mongoDeployment.Status.ReadyReplicas
	desiredReplicas := *mongoDeployment.Spec.Replicas
	if readyReplicas == desiredReplicas {
		state.State = "Ready"
	}

	return state
}

func getPostgresState(ctx context.Context, name string, id string, projectID string) model.OperatorState {
	state := model.OperatorState{
		Name:  id,
		State: "Failed",
	}
	sqlService, err := sqladmin.NewService(ctx, option.WithScopes())
	if err != nil {
		return state
	}
	splitName := strings.Split(name, "/")
	dbName := splitName[1] + "-" + id
	handle := splitName[0]
	log.Println("Trying to get status for CloudSQL database: ", dbName)

	// Retrieve the instance status
	list, err := sqlService.Instances.List(projectID).Context(ctx).Do()
	if err != nil {
		return state
	}
	for _, db := range list.Items {
		if db.Settings.UserLabels["kapeta-deploymentname"] == dbName && db.Settings.UserLabels["kapeta-handle"] == handle {
			state.State = stateMapper(db.State)
			fmt.Printf("Instance status: %s\n", db.State)
			// early return since there should only be one
			return state
		}
	}

	return state
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
