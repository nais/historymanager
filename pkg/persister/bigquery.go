package persister

import (
	"context"
	"os"

	"cloud.google.com/go/bigquery"
	"github.com/nais/historymanager/pkg/models"
)

func Persist(ctx context.Context, topics []models.BigQueryAlert) error {
	client, err := bigquery.NewClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		return err
	}

	tableHandle := client.Dataset("alert_history").Table("alerts")
	inserter := tableHandle.Inserter()
	return inserter.Put(ctx, topics)
}
