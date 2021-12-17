package v1alerts

import (
	"context"

	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type AlertInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*naisiov1.AlertList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*naisiov1.Alert, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type alertClient struct {
	restClient rest.Interface
}

func (c *alertClient) List(ctx context.Context, opts metav1.ListOptions) (*naisiov1.AlertList, error) {
	result := naisiov1.AlertList{}
	err := c.restClient.
		Get().
		Resource("alerts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *alertClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*naisiov1.Alert, error) {
	result := naisiov1.Alert{}
	err := c.restClient.
		Get().
		Resource("alerts").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *alertClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Resource("alerts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}
