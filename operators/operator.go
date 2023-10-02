package operators

import (
	"context"

	"github.com/kapetacom/insight-api/model"
	"github.com/kapetacom/insight-api/operators/gcp"
	"github.com/kapetacom/insight-api/operators/local"
	"k8s.io/client-go/kubernetes"
)

func GetDatabaseState(ctx context.Context, mode string, clientset *kubernetes.Clientset) ([]model.OperatorState, error) {
	if mode == "kubernetes-only" {
		return local.GetDatabaseState(ctx, clientset)
	}
	return gcp.GetDatabaseState(ctx, clientset)

}
