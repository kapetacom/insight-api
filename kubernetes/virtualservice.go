package kubernetes

import (
	"context"

	istionetworking "istio.io/client-go/pkg/apis/networking/v1beta1"
	is "istio.io/client-go/pkg/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func GetVirtualService(ctx context.Context, config *rest.Config, namespace string) ([]*istionetworking.VirtualService, error) {
	c, err := is.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	endpoint := c.NetworkingV1beta1().VirtualServices(namespace)
	gw, err := endpoint.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	result := []*istionetworking.VirtualService{}
	result = append(result, gw.Items...)

	return result, nil

}
