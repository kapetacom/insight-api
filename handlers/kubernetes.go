package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/kapetacom/insight-api/jwt"
	"github.com/kapetacom/insight-api/kubernetes"
	"github.com/kapetacom/insight-api/model"
	"github.com/kapetacom/insight-api/operators"
	"github.com/kapetacom/insight-api/scopes"
	"github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Routes) GetEnvironmentStatus(c echo.Context) error {
	// TODO: Verify current user has access proper to this cluster
	if !jwt.HasScopeForHandle(c, os.Getenv("KAPETA_HANDLE"), scopes.RUNTIME_READ_SCOPE) {
		return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("user does not have access to this deployment, missing scope %v for %v", scopes.RUNTIME_READ_SCOPE, os.Getenv("KAPETA_HANDLE")))

	}
	clientset, err := kubernetes.KubernetesClient()
	if err != nil {
		return fmt.Errorf("error getting kubernetes client: %v\n", err)
	}
	deployments, err := clientset.AppsV1().Deployments("services").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting deployments: %v\n", err)
	}

	clusterStatus, err := getEnvironmentInfo(context.Background())
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	result := []model.InstanceState{}
	for _, deployment := range deployments.Items {
		// Get the number of ready replicas and desired replicas
		readyReplicas := deployment.Status.ReadyReplicas
		desiredReplicas := *deployment.Spec.Replicas
		blockID := deployment.GetObjectMeta().GetLabels()["kapeta.com/block-id"]
		// Print the readiness status
		if readyReplicas == desiredReplicas {
			result = append(result, model.InstanceState{Name: deployment.Name, State: "Ready", ReadyReplicas: readyReplicas, DesiredReplicas: desiredReplicas, BlockID: blockID})
		} else {
			// not sure what to call this state yet
			result = append(result, model.InstanceState{Name: deployment.Name, State: "Failed", ReadyReplicas: readyReplicas, DesiredReplicas: desiredReplicas, BlockID: blockID})
		}
	}
	providers, err := operators.GetDatabaseState(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	clusterStatus.Operators = providers
	clusterStatus.Instances = result
	return c.JSON(200, clusterStatus)
}

func getEnvironmentInfo(ctx context.Context) (*model.ClusterStatus, error) {
	// set the label selector for the secret
	labelSelector := "kapeta.com/environment-name"

	clientset, err := kubernetes.KubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client: %v\n", err)
	}
	// get the secrets that match the label selector
	secrets, err := clientset.CoreV1().Secrets("kapeta").List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		panic(err.Error())
	}

	if len(secrets.Items) > 0 {
		status := &model.ClusterStatus{}
		secret := secrets.Items[0]
		fmt.Printf("Secret name: %s\n", secret.ObjectMeta.Name)
		status.EnvironmentName = secret.GetObjectMeta().GetLabels()["kapeta.com/environment-name"]
		status.EnvironmentVersion = secret.GetObjectMeta().GetLabels()["kapeta.com/environment-version"]

		status.PlanName = secret.GetObjectMeta().GetLabels()["kapeta.com/plan-name"]
		status.PlanVersion = secret.GetObjectMeta().GetLabels()["kapeta.com/plan-version"]

		status.TargetName = secret.GetObjectMeta().GetLabels()["kapeta.com/deployment-target-name"]
		status.TargetVersion = secret.GetObjectMeta().GetLabels()["kapeta.com/deployment-target-version"]
		return status, nil
	}
	return nil, fmt.Errorf("Not the corret number of secrets was expecting 1 got %v", len(secrets.Items))
}
