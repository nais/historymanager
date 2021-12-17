package v1alerts

import (
	"context"
	"time"

	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func WatchAlerts(ctx context.Context, clientSet AlertV1Interface) cache.Store {
	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(listOptions metav1.ListOptions) (result runtime.Object, err error) {
				return clientSet.Alerts().List(ctx, listOptions)
			},
			WatchFunc: func(listOptions metav1.ListOptions) (watch.Interface, error) {
				return clientSet.Alerts().Watch(ctx, listOptions)
			},
		},
		&naisiov1.Alert{},
		5*time.Minute,
		cache.ResourceEventHandlerFuncs{},
	)

	go controller.Run(wait.NeverStop)

	for !controller.HasSynced() {
		time.Sleep(1 * time.Second)
	}
	return store
}
