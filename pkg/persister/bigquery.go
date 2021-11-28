package persister

import (
	"context"

	"cloud.google.com/go/bigquery"
	"github.com/nais/historymanager/pkg/models"
)

const BigQueryProjectID = "aura-dev-d9f5"

func Persist(ctx context.Context, topics []models.BqAlert) error {
	client, err := bigquery.NewClient(ctx, BigQueryProjectID)
	if err != nil {
		return err
	}

	tableHandle := client.Dataset("alert_history").Table("alerts")
	inserter := tableHandle.Inserter()
	return inserter.Put(ctx, topics)
}
