package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/kapetacom/insight-api/jwt"
	"github.com/kapetacom/insight-api/kubernetes"
	"github.com/kapetacom/insight-api/model"
	"github.com/kapetacom/insight-api/operators"
	"github.com/kapetacom/insight-api/scopes"
	"github.com/labstack/echo/v4"
	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (h *Routes) GetEnvironmentStatus(c echo.Context) error {
	// TODO: Verify current user has access proper to this cluster
	if !jwt.HasScopeForHandle(c, os.Getenv("KAPETA_HANDLE"), scopes.RUNTIME_READ_SCOPE) {
		return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("user does not have access to this deployment, missing scope %v for %v", scopes.RUNTIME_READ_SCOPE, os.Getenv("KAPETA_HANDLE")))

	}
	clientset, err := kubernetes.KubernetesClient()
	if err != nil {
		return fmt.Errorf("error getting kubernetes client: %v", err)
	}
	deployments, err := clientset.AppsV1().Deployments("services").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting deployments: %v", err)
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

	gateways, err := h.GetIngress(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	clusterStatus.Operators = providers
	clusterStatus.Instances = append(result, gateways...)

	return c.JSON(200, clusterStatus)
}

// GetIngress returns the status of the ingress routes, by querying the traefik api
func (h *Routes) GetIngress(c echo.Context) ([]model.InstanceState, error) {
	result := []model.InstanceState{}
	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "http://traefik-dashboard-service.infrastructure.svc.cluster.local:8080"
	}
	if !jwt.HasScopeForHandle(c, os.Getenv("KAPETA_HANDLE"), scopes.RUNTIME_READ_SCOPE) {
		return nil, echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("user does not have access to this deployment, missing scope %v for %v", scopes.RUNTIME_READ_SCOPE, os.Getenv("KAPETA_HANDLE")))
	}

	clientset, err := kubernetes.DynamicKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client: %v", err)
	}

	ingressGVR := schema.GroupVersionResource{Group: "traefik.io", Version: "v1alpha1", Resource: "ingressroutes"}

	ingressroutes, err := clientset.Resource(ingressGVR).Namespace("services").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting ingress: %v", err)
	}

	for _, ingressItem := range ingressroutes.Items {

		ingress, err := getIngressRoutes(ingressItem)
		if err != nil {
			return nil, fmt.Errorf("error marshalling ingress: %v", err)
		}

		traefikRoutes, err := getRoutes(apiHost)
		if err != nil {
			return nil, fmt.Errorf("error getting routes from Traefik: %v", err)
		}
		for _, route := range *traefikRoutes {
			if strings.HasPrefix(route.Name, "services-"+ingress.GetName()) {
				traefikService, err := getService(apiHost, route.Service)
				if err != nil {
					return nil, fmt.Errorf("error getting service from Traefik: %v", err)
				}
				status := "Ready"
				for _, ready := range traefikService.ServerStatus {
					if ready != "UP" {
						status = "Failed"
					}
				}
				instanceId := ingress.GetObjectMeta().GetLabels()["kapeta.com/instanceid"]
				apiPath := ingress.GetObjectMeta().GetAnnotations()["kapeta.com/api_path"]
				result = append(result, model.InstanceState{
					Name:     instanceId,
					BlockID:  instanceId,
					State:    status,
					Metadata: map[string]string{"kapeta.com/api_path": apiPath},
				})
			}
		}
	}
	return result, nil
}

func getRoutes(apiHost string) (*model.TraefikRoutes, error) {
	resp, err := http.Get(apiHost + "/api/http/routers")
	if err != nil {
		return nil, fmt.Errorf("error getting routers: %v", err)
	}
	traefikRoutes := &model.TraefikRoutes{}
	err = json.NewDecoder(resp.Body).Decode(traefikRoutes)
	if err != nil {
		return nil, fmt.Errorf("error decoding routers: %v", err)
	}
	return traefikRoutes, nil
}

func getService(apiHost string, serviceName string) (*model.TraefikService, error) {
	url := apiHost + "/api/http/services/" + serviceName + "@kubernetescrd"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting routers: %v", err)
	}
	traefikService := &model.TraefikService{}
	err = json.NewDecoder(resp.Body).Decode(traefikService)
	if err != nil {
		return nil, fmt.Errorf("error decoding routers: %v", err)
	}
	return traefikService, nil
}

func getIngressRoutes(item unstructured.Unstructured) (*traefik.IngressRoute, error) {
	j, err := item.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshalling ingress: %v", err)
	}

	ingress := traefik.IngressRoute{}
	err = json.Unmarshal(j, &ingress)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ingress: %v", err)
	}
	return &ingress, nil
}

func getEnvironmentInfo(ctx context.Context) (*model.ClusterStatus, error) {
	// set the label selector for the secret
	labelSelector := "kapeta.com/environment-name"

	clientset, err := kubernetes.KubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client: %v", err)
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
	return nil, fmt.Errorf("not the corret number of secrets was expecting 1 got %v", len(secrets.Items))
}
