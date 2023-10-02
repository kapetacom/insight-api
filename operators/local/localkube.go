package local

import (
	"context"
	"log"

	"github.com/kapetacom/insight-api/kubernetes"
	"github.com/kapetacom/insight-api/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

func GetDatabaseState(ctx context.Context, clientset *kube.Clientset) ([]model.OperatorState, error) {

	deployment, err := kubernetes.GetDeployment(ctx)
	if err != nil {
		return nil, err
	}

	states := []model.OperatorState{}

	for _, svc := range deployment.Spec.Services {
		switch svc.Kind {
		case "kapeta/resource-type-postgresql":
			states = append(states, getLocalPostgresState(ctx, deployment.Metadata.Name, svc.Id, "CHANGE ME"))
		case "kapeta/resource-type-mongodb":
			states = append(states, getMongoDBState(ctx, svc.Id))
		}
	}

	return states, nil
}

func getLocalPostgresState(ctx context.Context, name string, id string, projectID string) model.OperatorState {
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
	pgDeployment := podList.Items[0]
	readyReplicas := pgDeployment.Status.ReadyReplicas
	desiredReplicas := *pgDeployment.Spec.Replicas
	if readyReplicas == desiredReplicas {
		state.State = "Ready"
	}
	state.Name = pgDeployment.Name
	return state
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
