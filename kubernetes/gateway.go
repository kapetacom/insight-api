package kubernetes

import (
	"context"

	istionetworking "istio.io/client-go/pkg/apis/networking/v1beta1"
	is "istio.io/client-go/pkg/clientset/versioned"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetGateways(ctx context.Context, config *rest.Config, namespace string) ([]*istionetworking.Gateway, error) {
	c, err := is.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	endpoint := c.NetworkingV1beta1().Gateways(namespace)
	gw, err := endpoint.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	result := []*istionetworking.Gateway{}
	result = append(result, gw.Items...)

	return result, nil

}
func CreateOrUpdateGateway(ctx context.Context, clientset *kubernetes.Clientset, config *rest.Config, gateway *istionetworking.Gateway) error {
	c, err := is.NewForConfig(config)
	if err != nil {
		return err
	}
	endpoint := c.NetworkingV1beta1().Gateways(gateway.Namespace)
	_, err = endpoint.Get(ctx, gateway.Name, metav1.GetOptions{})
	if err != nil {
		// If the cluster role doesn't exist, create it
		if errors.IsNotFound(err) {
			_, err = endpoint.Create(ctx, gateway, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		oldResource, err := endpoint.Get(ctx, gateway.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		gateway.ResourceVersion = oldResource.ResourceVersion
		_, err = endpoint.Update(ctx, gateway, metav1.UpdateOptions{TypeMeta: oldResource.TypeMeta})
		if err != nil {
			return err
		}
	}
	return nil
}
