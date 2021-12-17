package v1alerts

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type AlertV1Interface interface {
	Alerts() AlertInterface
}

type AlertV1Client struct {
	restClient rest.Interface
}

func NewForConfig(config *rest.Config) (*AlertV1Client, error) {
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: "nais.io", Version: "v1"}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		return nil, err
	}

	return &AlertV1Client{restClient: client}, nil
}

func (c *AlertV1Client) Alerts() AlertInterface {
	return &alertClient{
		restClient: c.restClient,
	}
}
