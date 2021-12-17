package models

import (
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	"k8s.io/client-go/tools/cache"
)

type Labels struct {
	Alert               string
	Alertname           string `json:"alertname"`
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

type BigQueryAlert struct {
	Alertname         string
	Receiver          string
	App               string
	Namespace         string
	TriggerdNamespace string                `bigquery:"triggered_namespace"`
	StartsAt          civil.DateTime        `bigquery:"starts_at"`
	EndsAt            bigquery.NullDateTime `bigquery:"ends_at"`
	Status            string
	Fingerprint       string
	Cluster           string
}

func getFromStoreByDashKey(store cache.Store, name string) (string, error) {
	for _, v := range store.ListKeys() {
		key := strings.Replace(v, "/", "-", 1)
		if key == name {
			item, _, err := store.GetByKey(v)

			if err != nil {
				return "", err
			}
			naisAlert := item.(*nais_io_v1.Alert)
			return naisAlert.Namespace, nil
		}
	}

	return "", nil
}

func (a *Alert) AsBigQueryAlert(store cache.Store) (BigQueryAlert, error) {
	startsAt, err := civil.ParseDateTime(a.StartsAt.Format("2006-01-02t15:04:05.999999999"))
	if err != nil {
		return BigQueryAlert{}, err
	}

	endsAt, err := civil.ParseDateTime(a.EndsAt.Format("2006-01-02t15:04:05.999999"))
	if err != nil {
		return BigQueryAlert{}, err
	}
	nullEndsAt := bigquery.NullDateTime{
		DateTime: endsAt,
		Valid:    true,
	}

	if endsAt.String() == "0001-01-01T00:00:00" {
		nullEndsAt.Valid = false
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

	triggeredNamespace := a.Labels.Namespace
	if triggeredNamespace == "" {
		triggeredNamespace = a.Labels.LogNamespace
	}
	if triggeredNamespace == "" {
		triggeredNamespace = a.Labels.LinkerdNamespace
	}
	if triggeredNamespace == "" {
		triggeredNamespace = a.Labels.KubernetesNamespace
	}

	namespace, err := getFromStoreByDashKey(store, a.Labels.Alert)
	if err != nil {
		return BigQueryAlert{}, err
	}

	return BigQueryAlert{
		Alertname:         a.Labels.Alertname,
		Receiver:          a.Labels.Alert,
		App:               app,
		Namespace:         namespace,
		TriggerdNamespace: triggeredNamespace,
		StartsAt:          startsAt,
		EndsAt:            nullEndsAt,
		Status:            a.Status,
		Fingerprint:       a.Fingerprint,
		Cluster:           os.Getenv("NAIS_CLUSTER_NAME"),
	}, nil
}

func (ar *AlertRequest) ToBigQuery(store cache.Store) ([]BigQueryAlert, error) {
	var out []BigQueryAlert
	for _, a := range ar.Alerts {
		toBQ, err := a.AsBigQueryAlert(store)
		if err != nil {
			return nil, err
		}

		out = append(out, toBQ)
	}

	return out, nil
}
