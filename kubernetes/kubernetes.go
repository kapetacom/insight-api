package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kapetacom/schemas/packages/go/model"
	"github.com/mitchellh/go-homedir"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetDeployment returns the deployment object from the kapeta secret in the cluster, this is the latest version deployed
func GetDeployment(ctx context.Context) (*model.Deployment, error) {
	labelSelector := "kapeta.com/environment-name"

	clientset, err := KubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client: %v\n", err)
	}
	// get the secrets that match the label selector
	secrets, err := clientset.CoreV1().Secrets("kapeta").List(ctx, v1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		panic(err.Error())
	}
	secret := secrets.Items[0]
	data := secret.Data["config"]

	deployment := &model.Deployment{}
	err = json.Unmarshal(data, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

// KubernetesClient returns a kubernetes client either from the cluster or from the local config if we are not running in a cluster
func KubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println("getting kubernetes config from local kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", home()+"/.kube/config")
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("getting kubernetes config from cluster")
	}

	kubectl, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kubectl, nil
}

func DynamicKubernetesClient() (*dynamic.DynamicClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println("getting kubernetes config from local kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", home()+"/.kube/config")
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("getting kubernetes config from cluster")
	}

	kubectl, err := dynamic.NewForConfig(config)
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
