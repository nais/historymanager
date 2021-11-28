package models

import (
	"time"

	"cloud.google.com/go/civil"
)

type Labels struct {
	Alert               string
	Name                string `json:"alertname"`
	Deployment          string `json:"deployment"`
	App                 string `json:"app"`
	LogApp              string `json:"log_app"`
	LogNamespace        string `json:"log_namespace"`
	KubernetesName      string `json:"kubernetes_name"`
	KubernetesNamespace string `json:"kubernetes_namespace"`
	LinkerdDeployment   string `json:"linkerd_io_proxy_deployment"`
	LinkerdNamespace    string `json:"linkerd_io_workload_ns"`
	Namespace           string `json:"namespace"`
}

type Alert struct {
	Status       string
	Labels       Labels
	Annotations  map[string]string
	StartsAt     time.Time
	EndsAt       time.Time
	GeneratorUrl string
	Fingerprint  string
}

type AlertRequest struct {
	Receiver          string
	Status            string
	Alerts            []Alert
	GroupLabels       map[string]string
	CommonLabels      map[string]string
	CommonAnnotations map[string]string
	ExternalUrl       string
	Version           string
	GroupKey          string
}

type BqAlert struct {
	Name        string
	Alert       string
	App         string
	Namespace   string
	StartsAt    civil.DateTime `bigquery:"starts_at"`
	EndsAt      civil.DateTime `bigquery:"ends_at"`
	Status      string
	Fingerprint string
}

func (a *Alert) AsBQAlert() (BqAlert, error) {
	startsAt, err := civil.ParseDateTime(a.StartsAt.Format("2006-01-02t15:04:05.999999999"))
	if err != nil {
		return BqAlert{}, err
	}

	endsAt, err := civil.ParseDateTime(a.EndsAt.Format("2006-01-02t15:04:05.999999999"))
	if err != nil {
		return BqAlert{}, err
	}

	app := a.Labels.App
	if app == "" {
		app = a.Labels.LogApp
	}
	if app == "" {
		app = a.Labels.LinkerdDeployment
	}
	if app == "" {
		app = a.Labels.Deployment
	}
	if app == "" {
		app = a.Labels.KubernetesName
	}

	namespace := a.Labels.Namespace
	if namespace == "" {
		namespace = a.Labels.LogNamespace
	}
	if namespace == "" {
		namespace = a.Labels.LinkerdNamespace
	}
	if namespace == "" {
		namespace = a.Labels.KubernetesNamespace
	}

	return BqAlert{
		Name:        a.Labels.Name,
		Alert:       a.Labels.Alert,
		App:         app,
		Namespace:   namespace,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Status:      a.Status,
		Fingerprint: a.Fingerprint,
	}, nil
}

func (ar *AlertRequest) ForBQ() ([]BqAlert, error) {
	var out []BqAlert
	for _, a := range ar.Alerts {
		toBQ, err := a.AsBQAlert()
		if err != nil {
			return nil, err
		}

		out = append(out, toBQ)
	}

	return out, nil
}
