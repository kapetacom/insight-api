package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kapetacom/insight-api/jwt"
	"github.com/kapetacom/insight-api/scopes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterStatus struct {
	Instances          []InstanceState `json:"instanceStates"`
	EnvironmentName    string          `json:"environmentName"`
	EnvironmentVersion string          `json:"environmentVersion"`
	PlanName           string          `json:"planName"`
	PlanVersion        string          `json:"planVersion"`
	TargetName         string          `json:"targetName"`
	TargetVersion      string          `json:"targetVersion"`
}

type InstanceState struct {
	Name            string `json:"name"`
	BlockID         string `json:"instanceId"`
	State           string `json:"state"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	DesiredReplicas int32  `json:"desiredReplicas"`
}

func (h *Routes) GetEnvironmentStatus(c echo.Context) error {
	// TODO: Verify current user has access proper to this cluster
	if !jwt.HasScopeForHandle(c, os.Getenv("KAPETA_HANDLE"), scopes.RUNTIME_READ_SCOPE) {
		return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("user does not have access to this deployment, missing scope %v for %v", scopes.RUNTIME_READ_SCOPE, os.Getenv("KAPETA_HANDLE")))

	}
	clientset, err := KubernetesClient()
	if err != nil {
		return fmt.Errorf("error getting kubernetes config: %v\n", err)
	}
	pods, err := clientset.AppsV1().Deployments("services").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting pods: %v\n", err)
	}

	clusterStatus, err := getEnvironmentInfo(context.Background(), clientset)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err)
	}

	result := []InstanceState{}
	for _, deployment := range pods.Items {
		// Get the number of ready replicas and desired replicas
		readyReplicas := deployment.Status.ReadyReplicas
		desiredReplicas := *deployment.Spec.Replicas
		blockID := deployment.GetObjectMeta().GetLabels()["kapeta.com/blockid"]
		// Print the readiness status
		if readyReplicas == desiredReplicas {
			result = append(result, InstanceState{Name: deployment.Name, State: "Ready", ReadyReplicas: readyReplicas, DesiredReplicas: desiredReplicas, BlockID: blockID})
		} else {
			// not sure what to call this state yet
			result = append(result, InstanceState{Name: deployment.Name, State: "Failed", ReadyReplicas: readyReplicas, DesiredReplicas: desiredReplicas, BlockID: blockID})
		}
	}
	clusterStatus.Instances = result
	return c.JSON(200, clusterStatus)
}

func getEnvironmentInfo(ctx context.Context, clientset *kubernetes.Clientset) (*ClusterStatus, error) {
	// set the label selector for the secret
	labelSelector := "kapeta.com/environment-name"

	// get the secrets that match the label selector
	secrets, err := clientset.CoreV1().Secrets("kapeta").List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		panic(err.Error())
	}

	status := &ClusterStatus{}
	if len(secrets.Items) > 0 {
		secret := secrets.Items[0]
		fmt.Printf("Secret name: %s\n", secret.ObjectMeta.Name)
		status.EnvironmentName = secret.GetObjectMeta().GetLabels()["kapeta.com/environment-name"]
		status.EnvironmentVersion = secret.GetObjectMeta().GetLabels()["kapeta.com/environment-version"]

		status.PlanName = secret.GetObjectMeta().GetLabels()["kapeta.com/plan-name"]
		status.PlanVersion = secret.GetObjectMeta().GetLabels()["kapeta.com/plan-version"]

		status.TargetName = secret.GetObjectMeta().GetLabels()["kapeta.com/target-name"]
		status.TargetVersion = secret.GetObjectMeta().GetLabels()["kapeta.com/target-version"]
	} else {
		return nil, fmt.Errorf("Not the corret number of secrets was expecting 1 got %v", len(secrets.Items))
	}
	return status, nil
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
