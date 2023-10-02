package gcp

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
	kube "k8s.io/client-go/kubernetes"
)

func GetDatabaseState(ctx context.Context, clientset *kube.Clientset) ([]model.OperatorState, error) {

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
			states = append(states, getMongoDBState(ctx, svc.Id))
		}
	}

	return states, nil
}

func getMongoDBState(ctx context.Context, id string) model.OperatorState {
	state := model.OperatorState{
		ID:    id,
		State: "Failed",
	}
	clientset, err := kubernetes.KubernetesClient()
	if err != nil {
		log.Printf("error getting kubernetes client: %v\n", err)
		return state
	}
	podList, err := clientset.AppsV1().Deployments("infrastructure").List(ctx, metav1.ListOptions{
		LabelSelector: "kapeta.com/block-id=" + id,
	})
	if err != nil {
		log.Printf("error getting pods: %v\n", err)
		return state
	}
	if len(podList.Items) == 0 {
		log.Printf("no pods found with label kapeta.com/block-id=%s\n", id)
		return state
	}
	mongoDeployment := podList.Items[0] //TODO: More than one pod?
	readyReplicas := mongoDeployment.Status.ReadyReplicas
	desiredReplicas := *mongoDeployment.Spec.Replicas
	if readyReplicas == desiredReplicas {
		state.State = "Ready"
	}
	state.Name = mongoDeployment.Name
	return state
}

func getPostgresState(ctx context.Context, name string, id string, projectID string) model.OperatorState {
	state := model.OperatorState{
		ID:    id,
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
	state.Name = dbName
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
