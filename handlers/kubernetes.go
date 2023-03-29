package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kapetacom/insight-api/jwt"
	"github.com/kapetacom/insight-api/scopes"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/go-homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type PodResult struct {
	Name            string `json:"name"`
	State           string `json:"state"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	DesiredReplicas int32  `json:"desiredReplicas"`
}

func (h *Routes) GetEnvironmentStatus(c echo.Context) error {
	// TODO: Verify current user has access proper to this cluster
	if !jwt.HasScopeForHandle(c, os.Getenv("KAPETA_HANDLE"), scopes.RUNTIME_READ_SCOPE) {
		return echo.NewHTTPError(http.StatusForbidden, "user does not have access to this deployment")
	}
	clientset, err := KubernetesClient()
	if err != nil {
		return fmt.Errorf("error getting kubernetes config: %v\n", err)
	}
	pods, err := clientset.AppsV1().Deployments("services").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting pods: %v\n", err)
	}
	result := []PodResult{}
	for _, deployment := range pods.Items {
		// Get the number of ready replicas and desired replicas
		readyReplicas := deployment.Status.ReadyReplicas
		desiredReplicas := *deployment.Spec.Replicas

		// Print the readiness status
		if readyReplicas == desiredReplicas {
			result = append(result, PodResult{Name: deployment.Name, State: "Ready", ReadyReplicas: readyReplicas, DesiredReplicas: desiredReplicas})
		} else {
			// not sure what to call this state yet
			result = append(result, PodResult{Name: deployment.Name, State: "Failed", ReadyReplicas: readyReplicas, DesiredReplicas: desiredReplicas})
		}
	}
	return c.JSON(200, result)
}

// KubernetesClient returns a kubernetes client either from the cluster or from the local config if we are not running in a cluster
func KubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println("Appears we are not running in a cluster")
		config, err = clientcmd.BuildConfigFromFlags("", Config())
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Seems like we are running in a Kubernetes cluster!!")
	}

	kubectl, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubectl, nil
}

func home() string {
	dir, err := homedir.Dir()
	if err != nil {
		panic(err.Error())
	}

	home, err := homedir.Expand(dir)
	if err != nil {
		panic(err.Error())
	}
	return home
}

func Config() string {
	return home() + "/.kube/config"
}
