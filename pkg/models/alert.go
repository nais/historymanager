package models

import (
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
)

type Labels struct {
	Alert          string
	Alertname      string `json:"alertname"`
	Namespace      string `json:"namespace"`
	Deployment     string `json:"deployment"`
	App            string `json:"app"`
	LogApp         string `json:"log_app"`
	KubernetesName string `json:"kubernetes_name"`
	Cluster        string `json:"tenant_cluster"`
}

type Alert struct {
	Status      string
	Labels      Labels
	StartsAt    time.Time
	EndsAt      time.Time
	Fingerprint string
}

type AlertRequest struct {
	Receiver string
	Status   string
	Alerts   []Alert
}

type BigQueryAlert struct {
	Alertname   string
	Receiver    string
	App         string
	Namespace   string
	StartsAt    civil.DateTime        `bigquery:"starts_at"`
	EndsAt      bigquery.NullDateTime `bigquery:"ends_at"`
	Status      string
	Fingerprint string
	Cluster     string
}

func (a *Alert) AsBigQueryAlert() (BigQueryAlert, error) {
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
		app = a.Labels.Deployment
	}
	if app == "" {
		app = a.Labels.KubernetesName
	}

	return BigQueryAlert{
		Alertname:   a.Labels.Alertname,
		Receiver:    a.Labels.Alert,
		App:         app,
		Namespace:   a.Labels.Namespace,
		StartsAt:    startsAt,
		EndsAt:      nullEndsAt,
		Status:      a.Status,
		Fingerprint: a.Fingerprint,
		Cluster:     a.Labels.Cluster,
	}, nil
}

func (ar *AlertRequest) ToBigQuery() ([]BigQueryAlert, error) {
	var out []BigQueryAlert
	for _, a := range ar.Alerts {
		toBQ, err := a.AsBigQueryAlert()
		if err != nil {
			return nil, err
		}

		out = append(out, toBQ)
	}

	return out, nil
}
